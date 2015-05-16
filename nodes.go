// Copyright (c) 2015 Serge Gebhardt. All rights reserved.
//
// Use of this source code is governed by the ISC
// license that can be found in the LICENSE file.

package acd

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"errors"
	"github.com/google/go-querystring/query"
)

// NodesService provides access to the nodes in the Amazon Cloud Drive API.
//
// See: https://developer.amazon.com/public/apis/experience/cloud-drive/content/nodes
type NodesService struct {
	client *Client
}

// Gets the root folder of the Amazon Cloud Drive.
func (s *NodesService) GetRoot() (*Folder, *http.Response, error) {
	opts := &NodeListOptions{Filters: "kind:FOLDER AND isRoot:true"}

	roots, resp, err := s.GetNodes(opts)
	if err != nil {
		return nil, resp, err
	}

	if len(roots) < 1 {
		return nil, resp, errors.New("No root found")
	}

	return &Folder{roots[0]}, resp, nil
}

// Gets the list of all nodes.
func (s *NodesService) GetAllNodes(opts *NodeListOptions) ([]*Node, *http.Response, error) {
	return s.listAllNodes("nodes", opts)
}

// Gets a list of nodes, up until the limit (either default or the one set in opts).
func (s *NodesService) GetNodes(opts *NodeListOptions) ([]*Node, *http.Response, error) {
	return s.listNodes("nodes", opts)
}

func (s *NodesService) listAllNodes(url string, opts *NodeListOptions) ([]*Node, *http.Response, error) {
	// Need opts to maintain state (NodeListOptions.reachedEnd)
	if opts == nil {
		opts = &NodeListOptions{}
	}

	result := make([]*Node, 0, 200)

	for {
		nodes, resp, err := s.listNodes(url, opts)
		if err != nil {
			return result, resp, err
		}
		if nodes == nil {
			break
		}

		result = append(result, nodes...)
	}

	return result, nil, nil
}

func (s *NodesService) listNodes(url string, opts *NodeListOptions) ([]*Node, *http.Response, error) {
	if opts.reachedEnd {
		return nil, nil, nil
	}

	url, err := addOptions(url, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	nodeList := &nodeListInternal{}
	resp, err := s.client.Do(req, nodeList)
	if err != nil {
		return nil, resp, err
	}

	if nodeList.NextToken != nil {
		opts.StartToken = *nodeList.NextToken
	} else {
		opts.reachedEnd = true
	}

	nodes := nodeList.Data
	// iterate over index since iterating over value would create a copy
	for i := range nodes {
		nodes[i].service = s
	}

	return nodes, resp, nil
}

type nodeListInternal struct {
	Count     *uint64 `json:"count"`
	NextToken *string `json:"nextToken"`
	Data      []*Node `json:"data"`
}

// Node represents a digital asset on the Amazon Cloud Drive, including files
// and folders, in a parent-child relationship. A node contains only metadata
// (e.g. folder) or it contains metadata and content (e.g. file).
type Node struct {
	Id   *string `json:"id"`
	Name *string `json:"name"`
	Kind *string `json:"kind"`

	service *NodesService
}

func (n *Node) Typed() interface{} {
	var result interface{}

	if n.Kind == nil {
		result = n
	} else {
		switch *n.Kind {
		case "FOLDER":
			result = &Folder{n}
		case "FILE":
			result = &File{n}
		default:
			result = n
		}
	}

	return result
}

// Represents a file and contains only metadata.
type File struct {
	*Node
}

// Represents a folder and contains only metadata.
type Folder struct {
	*Node
}

// Gets the list of all children.
func (f *Folder) GetAllChildren(opts *NodeListOptions) ([]*Node, *http.Response, error) {
	url := fmt.Sprintf("nodes/%s/children", *f.Id)
	return f.service.listAllNodes(url, opts)
}

// Gets a list of children, up until the limit (either default or the one set in opts).
func (f *Folder) GetChildren(opts *NodeListOptions) ([]*Node, *http.Response, error) {
	url := fmt.Sprintf("nodes/%s/children", *f.Id)
	return f.service.listNodes(url, opts)
}

// Gets the subfolder by name. It is an error if not exactly one subfolder is found.
func (f *Folder) GetFolder(name string) (*Folder, *http.Response, error) {
	n, resp, err := f.GetNode(name)
	if err != nil {
		return nil, resp, err
	}

	res, ok := n.Typed().(*Folder)
	if !ok {
		err := errors.New(fmt.Sprintf("Node '%s' is not a folder", name))
		return nil, resp, err
	}

	return res, resp, nil
}

// Gets the file by name. It is an error if not exactly one file is found.
func (f *Folder) GetFile(name string) (*File, *http.Response, error) {
	n, resp, err := f.GetNode(name)
	if err != nil {
		return nil, resp, err
	}

	res, ok := n.Typed().(*File)
	if !ok {
		err := errors.New(fmt.Sprintf("Node '%s' is not a file", name))
		return nil, resp, err
	}

	return res, resp, nil
}

// Gets the node by name. It is an error if not exactly one node is found.
func (f *Folder) GetNode(name string) (*Node, *http.Response, error) {
	filter := "parents:" + *f.Id + " AND name:" + name
	opts := &NodeListOptions{Filters: filter}

	nodes, resp, err := f.service.GetNodes(opts)
	if err != nil {
		return nil, resp, err
	}

	if len(nodes) < 1 {
		err := errors.New(fmt.Sprintf("No node '%s' found", name))
		return nil, resp, err
	}
	if len(nodes) > 1 {
		err := errors.New(fmt.Sprintf("Too many nodes '%s' found (%v)", name, len(nodes)))
		return nil, resp, err
	}

	return nodes[0], resp, nil
}

// WalkNodes walks the given node hierarchy, getting each node along the way, and returns
// the deepest node. If an error occurs, returns the furthest successful node and the list
// of HTTP responses.
func (f *Folder) WalkNodes(names ...string) (*Node, []*http.Response, error) {
	resps := make([]*http.Response, 0, len(names))

	if len(names) == 0 {
		return f.Node, resps, nil
	}

	// process each node except the last one
	fp := f
	for _, name := range names[:len(names)-1] {
		fn, resp, err := fp.GetFolder(name)
		resps = append(resps, resp)
		if err != nil {
			return fp.Node, resps, err
		}

		fp = fn
	}

	// process the last node
	nl, resp, err := fp.GetNode(names[len(names)-1])
	resps = append(resps, resp)
	if err != nil {
		return fp.Node, resps, err
	}

	return nl, resps, nil
}

// NodeListOptions holds the options when getting a list of nodes, such as the filter,
// sorting and pagination.
type NodeListOptions struct {
	Limit   uint   `url:"limit,omitempty"`
	Filters string `url:"filters,omitempty"`
	Sort    string `url:"sort,omitempty"`

	// Token where to start for next page (internal)
	StartToken string `url:"startToken,omitempty"`
	reachedEnd bool
}

// addOptions adds the parameters in opts as URL query parameters to s.  opts
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opts interface{}) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}
