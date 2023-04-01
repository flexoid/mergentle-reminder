package main

import (
	"testing"

	"github.com/flexoid/mergentle-reminder/mocks"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/xanzy/go-gitlab"
)

func TestFetchOpenedMergeRequests(t *testing.T) {
	config := &Config{
		Projects: []ConfigProject{
			{ID: 1},
			{ID: 2},
		},
	}

	mockGitLabClient := mocks.NewGitLabClient(t)

	options := gitlab.ListProjectMergeRequestsOptions{
		State:   gitlab.String("opened"),
		OrderBy: gitlab.String("updated_at"),
		Sort:    gitlab.String("desc"),
		WIP:     gitlab.String("no"),
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 50,
		},
	}

	mockGitLabClient.On("ListProjectMergeRequests", config.Projects[0].ID, &options).Return(
		[]*gitlab.MergeRequest{{IID: 1}},
		&gitlab.Response{CurrentPage: 1, NextPage: 2, TotalPages: 2},
		nil,
	).Once()

	optionsForPage2 := options
	optionsForPage2.Page = 2
	mockGitLabClient.On("ListProjectMergeRequests", config.Projects[0].ID, &optionsForPage2).Return(
		[]*gitlab.MergeRequest{{IID: 2}},
		&gitlab.Response{CurrentPage: 2, TotalPages: 2},
		nil,
	).Once()

	mockGitLabClient.On("ListProjectMergeRequests", config.Projects[1].ID, &options).Return(
		[]*gitlab.MergeRequest{{IID: 3}},
		&gitlab.Response{CurrentPage: 1, TotalPages: 1},
		nil,
	).Once()

	mockGitLabClient.On("GetMergeRequestApprovalsConfiguration", config.Projects[0].ID, 1).Return(
		&gitlab.MergeRequestApprovals{
			ApprovedBy: []*gitlab.MergeRequestApproverUser{
				{User: &gitlab.BasicUser{Name: "John Doe"}},
				{User: &gitlab.BasicUser{Name: "Jane Doe"}},
			},
		},
		&gitlab.Response{},
		nil,
	).Once()

	mockGitLabClient.On("GetMergeRequestApprovalsConfiguration", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(
		&gitlab.MergeRequestApprovals{
			ApprovedBy: []*gitlab.MergeRequestApproverUser{},
		},
		&gitlab.Response{},
		nil,
	).Twice()

	mrs, err := fetchOpenedMergeRequests(config, mockGitLabClient)

	mockGitLabClient.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(mrs))

	assert.Equal(t, 1, mrs[0].MergeRequest.IID)
	assert.Equal(t, 2, len(mrs[0].ApprovedBy))
	assert.Equal(t, "John Doe", mrs[0].ApprovedBy[0])
	assert.Equal(t, "Jane Doe", mrs[0].ApprovedBy[1])

	assert.Equal(t, 2, mrs[1].MergeRequest.IID)
	assert.Equal(t, 0, len(mrs[1].ApprovedBy))

	assert.Equal(t, 3, mrs[2].MergeRequest.IID)
	assert.Equal(t, 0, len(mrs[2].ApprovedBy))
}

func TestFetchProjectsFromGroups(t *testing.T) {
	config := &Config{
		Groups: []ConfigGroup{
			{ID: 1},
			{ID: 2},
		},
	}

	mockGitLabClient := mocks.NewGitLabClient(t)

	options := gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 50,
			Page:    1,
		},
	}

	mockGitLabClient.On("ListGroupProjects", config.Groups[0].ID, &options).Return(
		[]*gitlab.Project{
			{ID: 1},
			{ID: 2},
		},
		&gitlab.Response{CurrentPage: 1, NextPage: 2, TotalPages: 2},
		nil,
	).Once()

	optionsForPage2 := options
	optionsForPage2.Page = 2
	mockGitLabClient.On("ListGroupProjects", config.Groups[0].ID, &optionsForPage2).Return(
		[]*gitlab.Project{
			{ID: 3},
			{ID: 4},
		},
		&gitlab.Response{CurrentPage: 2, TotalPages: 2},
		nil,
	).Once()

	mockGitLabClient.On("ListGroupProjects", config.Groups[1].ID, &options).Return(
		[]*gitlab.Project{
			{ID: 5},
		},
		&gitlab.Response{CurrentPage: 1, TotalPages: 1},
		nil,
	).Once()

	projectIDs, err := fetchProjectsFromGroups(config, mockGitLabClient)

	mockGitLabClient.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(projectIDs))

	assert.Equal(t, 1, projectIDs[0])
	assert.Equal(t, 2, projectIDs[1])
	assert.Equal(t, 3, projectIDs[2])
	assert.Equal(t, 4, projectIDs[3])
	assert.Equal(t, 5, projectIDs[4])
}
