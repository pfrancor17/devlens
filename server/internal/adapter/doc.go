// Package adapter groups the driven adapters that implement the interfaces in
// package port: GitHub API, ONNX runtime, SQL store and vector index. Each one
// lives in its own subpackage and can be swapped without touching the domain.
package adapter
