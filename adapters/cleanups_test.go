package adapters

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
	"time"
)

func TestDeploymentInfos_AsIDs(t *testing.T) {
	var a deploymentInfos
	assert.Empty(t, a.AsIDs())

	a = append(a, deploymentInfo{
		id: "a",
		ts: time.Time{},
	})
	a = append(a, deploymentInfo{
		id: "b",
		ts: time.Time{},
	})
	a = append(a, deploymentInfo{
		id: "c",
		ts: time.Time{},
	})

	assert.ElementsMatch(t, a.AsIDs(), []string{"a", "b", "c"})
}

func TestDeploymentInfos_Sorting(t *testing.T) {
	testCases := []struct {
		name     string
		in       deploymentInfos
		expected deploymentInfos
	}{
		{
			name: "simple",
			in: deploymentInfos{
				{
					ts: time.Date(2012, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2014, 12, 1, 6, 32, 12, 0, time.UTC),
				},
			},
			expected: deploymentInfos{
				{
					ts: time.Date(2012, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2014, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 0, time.UTC),
				},
			},
		},
		{
			name: "simple2",
			in: deploymentInfos{
				{
					ts: time.Date(2016, 12, 1, 6, 32, 12, 2, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 3, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 1, time.UTC),
				},
			},
			expected: deploymentInfos{
				{
					ts: time.Date(2016, 12, 1, 6, 32, 12, 1, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 2, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 3, time.UTC),
				},
			},
		},
		{
			name:     "simple3",
			in:       deploymentInfos{},
			expected: deploymentInfos{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sorted := make(deploymentInfos, len(tc.in))
			copy(sorted, tc.in)
			sort.Sort(sorted)
			assert.Equal(t, tc.expected, sorted)
		})
	}
}

func TestGetDeletableDeployments(t *testing.T) {
	testCases := []struct {
		name       string
		in         deploymentInfos
		maxHistory uint
		expected   deploymentInfos
	}{
		{
			name: "simple_one",
			in: deploymentInfos{
				{
					ts: time.Date(2012, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2014, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 0, time.UTC),
				},
			},
			maxHistory: 2,
			expected: deploymentInfos{
				{
					ts: time.Date(2012, 12, 1, 6, 32, 12, 0, time.UTC),
				},
			},
		},
		{
			name: "simple_two",
			in: deploymentInfos{
				{
					ts: time.Date(2012, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2014, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 0, time.UTC),
				},
			},
			maxHistory: 1,
			expected: deploymentInfos{
				{
					ts: time.Date(2014, 12, 1, 6, 32, 12, 0, time.UTC),
				},
				{
					ts: time.Date(2012, 12, 1, 6, 32, 12, 0, time.UTC),
				},
			},
		},
		{
			name: "simple_three",
			in: deploymentInfos{
				{
					ts: time.Date(2012, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2014, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 0, time.UTC),
				},
			},
			maxHistory: 0,
			expected: deploymentInfos{
				{
					ts: time.Date(2012, 12, 1, 6, 32, 12, 0, time.UTC),
				},
				{
					ts: time.Date(2014, 12, 1, 6, 32, 12, 0, time.UTC),
				},
				{
					ts: time.Date(2016, 12, 1, 6, 32, 12, 0, time.UTC),
				},
			},
		},
		{
			name: "simple_none",
			in: deploymentInfos{
				{
					ts: time.Date(2012, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2014, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 0, time.UTC),
				},
			},
			maxHistory: 3,
			expected:   deploymentInfos{},
		},
		{
			name: "simple_none2",
			in: deploymentInfos{
				{
					ts: time.Date(2012, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2014, 12, 1, 6, 32, 12, 0, time.UTC),
				}, {
					ts: time.Date(2016, 12, 1, 6, 32, 12, 0, time.UTC),
				},
			},
			maxHistory: 25,
			expected:   deploymentInfos{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			toDelete := getDeletableDeployments(tc.maxHistory, tc.in)
			assert.ElementsMatch(t, tc.expected, toDelete)
		})
	}
}
