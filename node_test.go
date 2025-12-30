package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNode_AddAnchor(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.AddAnchor("test1")
	require.Equal(t, []string{"test1"}, node.Anchor)

	node.AddAnchor("test1")
	require.Equal(t, []string{"test1"}, node.Anchor)

	node.AddAnchor("test2")
	require.Equal(t, []string{"test1", "test2"}, node.Anchor)
}

func TestNode_RemoveAnchor(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.SetAnchor([]string{"test1", "test2", "test3", "test4"})
	require.Equal(t, []string{"test1", "test2", "test3", "test4"}, node.Anchor)

	node.RemoveAnchor("test2")
	require.Equal(t, []string{"test1", "test3", "test4"}, node.Anchor)

	node.RemoveAnchor("test2")
	require.Equal(t, []string{"test1", "test3", "test4"}, node.Anchor)

	node.RemoveAnchor("test4")
	require.Equal(t, []string{"test1", "test3"}, node.Anchor)
}

func TestNode_AddNext(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.AddNext("test1")
	require.Equal(t, []NodeNextItem{{Name: "test1"}}, node.Next)

	node.AddNext("test1")
	require.Equal(t, []NodeNextItem{{Name: "test1"}}, node.Next)

	node.AddNext("test2")
	require.Equal(t, []NodeNextItem{{Name: "test1"}, {Name: "test2"}}, node.Next)
}

func TestNode_RemoveNext(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.SetNext([]NodeNextItem{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}, {Name: "test4"}})
	require.Equal(t, []NodeNextItem{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}, {Name: "test4"}}, node.Next)

	node.RemoveNext("test2")
	require.Equal(t, []NodeNextItem{{Name: "test1"}, {Name: "test3"}, {Name: "test4"}}, node.Next)

	node.RemoveNext("test2")
	require.Equal(t, []NodeNextItem{{Name: "test1"}, {Name: "test3"}, {Name: "test4"}}, node.Next)

	node.RemoveNext("test4")
	require.Equal(t, []NodeNextItem{{Name: "test1"}, {Name: "test3"}}, node.Next)
}

func TestNode_AddOnError(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.AddOnError("test1")
	require.Equal(t, []NodeNextItem{{Name: "test1"}}, node.OnError)

	node.AddOnError("test1")
	require.Equal(t, []NodeNextItem{{Name: "test1"}}, node.OnError)

	node.AddOnError("test2")
	require.Equal(t, []NodeNextItem{{Name: "test1"}, {Name: "test2"}}, node.OnError)
}

func TestNode_RemoveOnError(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.SetOnError([]NodeNextItem{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}, {Name: "test4"}})
	require.Equal(t, []NodeNextItem{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}, {Name: "test4"}}, node.OnError)

	node.RemoveOnError("test2")
	require.Equal(t, []NodeNextItem{{Name: "test1"}, {Name: "test3"}, {Name: "test4"}}, node.OnError)

	node.RemoveOnError("test2")
	require.Equal(t, []NodeNextItem{{Name: "test1"}, {Name: "test3"}, {Name: "test4"}}, node.OnError)

	node.RemoveOnError("test4")
	require.Equal(t, []NodeNextItem{{Name: "test1"}, {Name: "test3"}}, node.OnError)
}
