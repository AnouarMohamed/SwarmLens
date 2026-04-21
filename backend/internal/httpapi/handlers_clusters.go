package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
	"github.com/AnouarMohamed/swarmlens/backend/internal/store"
)

func (d *deps) handleClustersList(w http.ResponseWriter, r *http.Request) {
	clusters, err := d.clustersWithHealth(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cluster_list_failed", err.Error())
		return
	}
	writeList(w, clusters, len(clusters))
}

func (d *deps) handleClustersGet(w http.ResponseWriter, r *http.Request) {
	clusterID := r.PathValue("clusterID")
	cluster, runtime, err := d.runtimeForCluster(r.Context(), clusterID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNotFound) {
			status = http.StatusNotFound
		}
		writeError(w, status, "cluster_get_failed", err.Error())
		return
	}
	cluster.Health = d.clusterHealth(runtime)
	writeOK(w, cluster)
}

func (d *deps) handleClustersCreate(w http.ResponseWriter, r *http.Request) {
	cluster, err := d.decodeClusterBody(r, "")
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	saved, err := d.store.SaveCluster(r.Context(), cluster)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cluster_create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": saved})
}

func (d *deps) handleClustersUpdate(w http.ResponseWriter, r *http.Request) {
	clusterID := r.PathValue("clusterID")
	cluster, err := d.decodeClusterBody(r, clusterID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	saved, err := d.store.SaveCluster(r.Context(), cluster)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cluster_update_failed", err.Error())
		return
	}
	writeOK(w, saved)
}

func (d *deps) decodeClusterBody(r *http.Request, existingID string) (model.Cluster, error) {
	current := model.Cluster{
		ID:             existingID,
		ConnectionMode: model.ClusterConnectionDirect,
		Enabled:        true,
	}
	if existingID != "" {
		existing, err := d.store.GetCluster(r.Context(), existingID)
		if err == nil {
			current = existing
		}
	}

	var body struct {
		Name           string `json:"name"`
		DockerHost     string `json:"dockerHost"`
		ConnectionMode string `json:"connectionMode"`
		TLSEnabled     *bool  `json:"tlsEnabled"`
		CertRef        string `json:"certRef"`
		Enabled        *bool  `json:"enabled"`
		Default        *bool  `json:"default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return model.Cluster{}, err
	}
	if strings.TrimSpace(body.Name) != "" {
		current.Name = strings.TrimSpace(body.Name)
	}
	if strings.TrimSpace(body.DockerHost) != "" {
		current.DockerHost = strings.TrimSpace(body.DockerHost)
	}
	if strings.TrimSpace(body.ConnectionMode) != "" {
		current.ConnectionMode = model.ClusterConnectionMode(strings.TrimSpace(body.ConnectionMode))
	}
	if body.TLSEnabled != nil {
		current.TLSEnabled = *body.TLSEnabled
	}
	if strings.TrimSpace(body.CertRef) != "" {
		current.CertRef = strings.TrimSpace(body.CertRef)
	}
	if body.Enabled != nil {
		current.Enabled = *body.Enabled
	}
	if body.Default != nil {
		current.Default = *body.Default
	}
	if current.Name == "" {
		return model.Cluster{}, errBadRequest("name is required")
	}
	if current.ConnectionMode == "" {
		current.ConnectionMode = model.ClusterConnectionDirect
	}
	if current.ConnectionMode != model.ClusterConnectionDemo && current.DockerHost == "" {
		return model.Cluster{}, errBadRequest("dockerHost is required for direct clusters")
	}
	if current.ConnectionMode == model.ClusterConnectionDemo {
		current.DockerHost = "demo"
		current.TLSEnabled = false
		current.CertRef = ""
	}
	if current.Default {
		current.Enabled = true
	}
	return current, nil
}

type badRequestError string

func (e badRequestError) Error() string { return string(e) }

func errBadRequest(message string) error {
	return badRequestError(message)
}
