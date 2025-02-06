package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestFilterMergeRequestsByAuthor(t *testing.T) {
	mrs := []*MergeRequestWithApprovals{
		{
			MergeRequest: &gitlab.MergeRequest{
				Author: &gitlab.BasicUser{ID: 1, Name: "John Doe", Username: "johndoe"},
			},
		},
		{
			MergeRequest: &gitlab.MergeRequest{
				Author: &gitlab.BasicUser{ID: 2, Name: "James Doe", Username: "jamesdoe"},
			},
		},
		{
			MergeRequest: &gitlab.MergeRequest{
				Author: &gitlab.BasicUser{ID: 2, Name: "Jane Doe", Username: "janedoe"},
			},
		},
	}

	authors := []ConfigAuthor{
		{ID: 1},
		{Username: "janedoe"},
	}

	filteredMRs := filterMergeRequestsByAuthor(mrs, authors)

	require.Equal(t, 2, len(filteredMRs))
	assert.Equal(t, 1, filteredMRs[0].MergeRequest.Author.ID)
	assert.Equal(t, 2, filteredMRs[1].MergeRequest.Author.ID)
}

func TestFilterMergeRequestsByAuthor_NoMatchingAuthors(t *testing.T) {
	mrs := []*MergeRequestWithApprovals{
		{
			MergeRequest: &gitlab.MergeRequest{
				Author: &gitlab.BasicUser{ID: 3, Name: "Alice", Username: "alice"},
			},
		},
	}
	authors := []ConfigAuthor{
		{ID: 1},
		{Username: "janedoe"},
	}
	filteredMRs := filterMergeRequestsByAuthor(mrs, authors)
	require.Equal(t, 0, len(filteredMRs))
}

func TestFilterMergeRequestsByAuthor_MultipleMergeRequestsSameAuthor(t *testing.T) {
	mrs := []*MergeRequestWithApprovals{
		{
			MergeRequest: &gitlab.MergeRequest{
				Author: &gitlab.BasicUser{ID: 1, Name: "John Doe", Username: "johndoe"},
			},
		},
		{
			MergeRequest: &gitlab.MergeRequest{
				Author: &gitlab.BasicUser{ID: 1, Name: "John Doe", Username: "johndoe"},
			},
		},
	}
	authors := []ConfigAuthor{
		{ID: 1},
	}
	filteredMRs := filterMergeRequestsByAuthor(mrs, authors)
	require.Equal(t, 2, len(filteredMRs))
	assert.Equal(t, 1, filteredMRs[0].MergeRequest.Author.ID)
	assert.Equal(t, 1, filteredMRs[1].MergeRequest.Author.ID)
}

func TestFilterMergeRequestsByAuthor_EmptyMergeRequests(t *testing.T) {
	mrs := []*MergeRequestWithApprovals{}
	authors := []ConfigAuthor{
		{ID: 1},
	}
	filteredMRs := filterMergeRequestsByAuthor(mrs, authors)
	require.Equal(t, 0, len(filteredMRs))
}

func TestFilterMergeRequestsByAuthor_OptionalAuthors(t *testing.T) {
	mrs := []*MergeRequestWithApprovals{
		{
			MergeRequest: &gitlab.MergeRequest{
				Author: &gitlab.BasicUser{ID: 1, Name: "John Doe", Username: "johndoe"},
			},
		},
	}
	authors := []ConfigAuthor{}
	filteredMRs := filterMergeRequestsByAuthor(mrs, authors)
	require.Equal(t, 1, len(filteredMRs))
}
