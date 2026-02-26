package maa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNode_AddAnchor(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.AddAnchor("test1")
	require.Equal(t, map[string]string{"test1": "test"}, node.Anchor)

	node.AddAnchor("test1")
	require.Equal(t, map[string]string{"test1": "test"}, node.Anchor)

	node.AddAnchor("test2")
	require.Equal(t, map[string]string{"test1": "test", "test2": "test"}, node.Anchor)
}

func TestNode_RemoveAnchor(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.SetAnchor(map[string]string{
		"test1": "test",
		"test2": "test",
		"test3": "test",
		"test4": "test",
	})
	require.Equal(t, map[string]string{
		"test1": "test",
		"test2": "test",
		"test3": "test",
		"test4": "test",
	}, node.Anchor)

	node.RemoveAnchor("test2")
	require.Equal(t, map[string]string{
		"test1": "test",
		"test3": "test",
		"test4": "test",
	}, node.Anchor)

	node.RemoveAnchor("test2")
	require.Equal(t, map[string]string{
		"test1": "test",
		"test3": "test",
		"test4": "test",
	}, node.Anchor)

	node.RemoveAnchor("test4")
	require.Equal(t, map[string]string{
		"test1": "test",
		"test3": "test",
	}, node.Anchor)
}

func TestNode_SetAnchorTarget(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.SetAnchorTarget("anchorA", "NodeA")
	node.SetAnchorTarget("anchorB", "NodeB")
	require.Equal(t, map[string]string{
		"anchorA": "NodeA",
		"anchorB": "NodeB",
	}, node.Anchor)
}

func TestNode_ClearAnchor(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.AddAnchor("anchorA")
	node.ClearAnchor("anchorA")
	require.Equal(t, map[string]string{
		"anchorA": "",
	}, node.Anchor)
}

func TestNode_AddNext(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.AddNext("test1")
	require.Equal(t, []NextItem{{Name: "test1"}}, node.Next)

	node.AddNext("test1")
	require.Equal(t, []NextItem{{Name: "test1"}}, node.Next)

	node.AddNext("test2")
	require.Equal(t, []NextItem{{Name: "test1"}, {Name: "test2"}}, node.Next)
}

func TestNode_RemoveNext(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.SetNext([]NextItem{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}, {Name: "test4"}})
	require.Equal(t, []NextItem{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}, {Name: "test4"}}, node.Next)

	node.RemoveNext("test2")
	require.Equal(t, []NextItem{{Name: "test1"}, {Name: "test3"}, {Name: "test4"}}, node.Next)

	node.RemoveNext("test2")
	require.Equal(t, []NextItem{{Name: "test1"}, {Name: "test3"}, {Name: "test4"}}, node.Next)

	node.RemoveNext("test4")
	require.Equal(t, []NextItem{{Name: "test1"}, {Name: "test3"}}, node.Next)
}

func TestNode_AddOnError(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.AddOnError("test1")
	require.Equal(t, []NextItem{{Name: "test1"}}, node.OnError)

	node.AddOnError("test1")
	require.Equal(t, []NextItem{{Name: "test1"}}, node.OnError)

	node.AddOnError("test2")
	require.Equal(t, []NextItem{{Name: "test1"}, {Name: "test2"}}, node.OnError)
}

func TestNode_RemoveOnError(t *testing.T) {
	node := NewNode("test")
	require.NotNil(t, node)

	node.SetOnError([]NextItem{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}, {Name: "test4"}})
	require.Equal(t, []NextItem{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}, {Name: "test4"}}, node.OnError)

	node.RemoveOnError("test2")
	require.Equal(t, []NextItem{{Name: "test1"}, {Name: "test3"}, {Name: "test4"}}, node.OnError)

	node.RemoveOnError("test2")
	require.Equal(t, []NextItem{{Name: "test1"}, {Name: "test3"}, {Name: "test4"}}, node.OnError)

	node.RemoveOnError("test4")
	require.Equal(t, []NextItem{{Name: "test1"}, {Name: "test3"}}, node.OnError)
}
