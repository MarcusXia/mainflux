// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package things

import (
	"context"

	"github.com/MainfluxLabs/mainflux/pkg/errors"
)

var (
	// ErrConnect indicates error in adding connection
	ErrConnect = errors.New("add connection failed")

	// ErrDisconnect indicates error in removing connection
	ErrDisconnect = errors.New("remove connection failed")

	// ErrEntityConnected indicates error while checking connection in database
	ErrEntityConnected = errors.New("check thing-channel connection in database error")
)

// Metadata to be used for Mainflux thing or channel for customized
// describing of particular thing or channel.
type Metadata map[string]interface{}

// Thing represents a Mainflux thing. Each thing is owned by one user, and
// it is assigned with the unique identifier and (temporary) access key.
type Thing struct {
	ID       string
	GroupID  string
	Name     string
	Key      string
	Metadata Metadata
}

// ThingsPage contains page related metadata as well as list of things that
// belong to this page.
type ThingsPage struct {
	PageMetadata
	Things []Thing
}

// ThingRepository specifies a thing persistence API.
type ThingRepository interface {
	// Save persists multiple things. Things are saved using a transaction. If one thing
	// fails then none will be saved. Successful operation is indicated by non-nil
	// error response.
	Save(ctx context.Context, ths ...Thing) ([]Thing, error)

	// Update performs an update to the existing thing. A non-nil error is
	// returned to indicate operation failure.
	Update(ctx context.Context, t Thing) error

	// UpdateKey updates key value of the existing thing. A non-nil error is
	// returned to indicate operation failure.
	UpdateKey(ctx context.Context, id, key string) error

	// RetrieveByID retrieves the thing having the provided identifier, that is owned
	// by the specified user.
	RetrieveByID(ctx context.Context, id string) (Thing, error)

	// RetrieveByKey returns thing ID for given thing key.
	RetrieveByKey(ctx context.Context, key string) (string, error)

	// RetrieveByGroupIDs retrieves the subset of things specified by given group ids.
	RetrieveByGroupIDs(ctx context.Context, groupIDs []string, pm PageMetadata) (ThingsPage, error)

	// RetrieveByChannel retrieves the subset of things owned by the specified
	// user and connected or not connected to specified channel.
	RetrieveByChannel(ctx context.Context, chID string, pm PageMetadata) (ThingsPage, error)

	// Remove removes the things having the provided identifiers, that is owned
	// by the specified user.
	Remove(ctx context.Context, ids ...string) error

	// RetrieveAll retrieves all things for all users.
	RetrieveAll(ctx context.Context) ([]Thing, error)

	// RetrieveByAdmin retrieves all things for all users with pagination.
	RetrieveByAdmin(ctx context.Context, pm PageMetadata) (ThingsPage, error)
}

// ThingCache contains thing caching interface.
type ThingCache interface {
	// Save stores pair thing key, thing id.
	Save(context.Context, string, string) error

	// ID returns thing ID for given key.
	ID(context.Context, string) (string, error)

	// Remove removes thing from cache.
	Remove(context.Context, string) error

	// SaveRole stores pair groupID:memberID, role.
	SaveRole(context.Context, string, string, string) (error)

	// Role stores pair groupID:memberID, role.
	Role(context.Context, string, string) (string, error)

	// RemoveRole removes group member role from cache.
	RemoveRole(context.Context, string, string) (error)
}
