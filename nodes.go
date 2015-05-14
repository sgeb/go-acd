// Copyright (c) 2015 Serge Gebhardt. All rights reserved.
//
// Use of this source code is governed by the ISC
// license that can be found in the LICENSE file.

package acd

import (
	"net/http"
)

// NodesService provides access to the nodes in the Amazon Cloud Drive API.
//
// See: https://developer.amazon.com/public/apis/experience/cloud-drive/content/nodes
type NodesService struct {
	client *Client
}

type nodeListInternal struct {
	Count     *uint64 `json: "count"`
	NextToken *string `json: "nextToken"`
	Data      []Node  `json: "data"`
}

// Node represents the different a digital asset on the Amazon Cloud Drive, including files
// and folders, in a parent-child relationship. A node contains only metadata (e.g. folder)
// or it contains metadata and content (e.g. file).
type Node struct {
	Id   *string `json: "id"`
	Name *string `json:"name"`
}

// NodeListOptions holds the options when getting a list nodes, such as the filter, sorting
// and pagination.
type NodeListOptions struct {
	NextPageToken string
}

// Gets a list of nodes.
func (s *NodesService) GetNodes(opts *NodeListOptions) ([]Node, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "nodes", nil)
	if err != nil {
		return nil, nil, err
	}

	nodeList := &nodeListInternal{}
	resp, err := s.client.Do(req, nodeList)
	if err != nil {
		return nil, resp, err
	}

	opts.NextPageToken = *nodeList.NextToken
	return nodeList.Data, resp, err
}
