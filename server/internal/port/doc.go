// Package port declares the driven ports of DevLens: the interfaces use cases
// need from the outside world (code repositories, model predictors, result
// stores, vector indexes). Adapters implement them; use cases depend on them.
//
// Ports are added together with the first use case that needs them, never
// ahead of time.
package port
