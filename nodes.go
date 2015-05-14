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

	return &Folder{&roots[0]}, resp, nil
}

// Gets the list of all nodes.
func (s *NodesService) GetAllNodes(opts *NodeListOptions) ([]Node, *http.Response, error) {
	return s.listAllNodes("nodes", opts)
}

// Gets a list of nodes, up until the limit (either default or the one set in opts).
func (s *NodesService) GetNodes(opts *NodeListOptions) ([]Node, *http.Response, error) {
	return s.listNodes("nodes", opts)
}

func (s *NodesService) listAllNodes(url string, opts *NodeListOptions) ([]Node, *http.Response, error) {
	// Need opts to maintain state (NodeListOptions.reachedEnd)
	if opts == nil {
		opts = &NodeListOptions{}
	}

	result := make([]Node, 0, 200)

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

func (s *NodesService) listNodes(url string, opts *NodeListOptions) ([]Node, *http.Response, error) {
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
	Data      []Node  `json:"data"`
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
func (f *Folder) GetAllChildren(opts *NodeListOptions) ([]Node, *http.Response, error) {
	url := fmt.Sprintf("nodes/%s/children", *f.Id)
	return f.service.listAllNodes(url, opts)
}

// Gets a list of children, up until the limit (either default or the one set in opts).
func (f *Folder) GetChildren(opts *NodeListOptions) ([]Node, *http.Response, error) {
	url := fmt.Sprintf("nodes/%s/children", *f.Id)
	return f.service.listNodes(url, opts)
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
