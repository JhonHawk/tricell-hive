// model_catalog_cache.go keeps the model lists the Models view has asked for
// during one run of the application (#46, T9). appModel creates one cache and
// every view reaches it through appConfig, which is copied by value, so the
// pointer is the part that is shared. A list is stored by appModel.Update when
// the result message arrives, a message no view owns, so a query that finishes
// after the person left the Models view is not lost. A failure is never kept:
// the next time a panel opens, the query runs again.
package main

// modelCatalogCache holds the lists by CLI and which queries are in flight.
type modelCatalogCache struct {
	models  map[string][]catalogModel
	loading map[string]bool
}

func newModelCatalogCache() *modelCatalogCache {
	return &modelCatalogCache{models: map[string][]catalogModel{}, loading: map[string]bool{}}
}

// get returns the stored list of host, when one was loaded.
func (c *modelCatalogCache) get(host string) ([]catalogModel, bool) {
	models, ok := c.models[host]
	return models, ok
}

func (c *modelCatalogCache) isLoading(host string) bool { return c.loading[host] }

// begin marks a query of host as running and reports whether the caller must
// start it: false when the list is stored or a query is already in flight.
func (c *modelCatalogCache) begin(host string) bool {
	if _, ok := c.models[host]; ok || c.loading[host] {
		return false
	}
	c.loading[host] = true
	return true
}

// finish records a query's result. Only appModel.Update calls it.
func (c *modelCatalogCache) finish(host string, models []catalogModel, err error) {
	delete(c.loading, host)
	if err == nil {
		c.models[host] = models
	}
}

// catalogLoadedMsg is the result of one CLI's model query. It carries the cache
// it was started for and is not owned by a view.
type catalogLoadedMsg struct {
	cache  *modelCatalogCache
	host   string
	models []catalogModel
	err    error
}
