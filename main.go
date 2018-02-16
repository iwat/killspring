package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/mitchellh/go-ps"
)

type _TreeNode struct {
	proc     ps.Process
	children []*_TreeNode
}

func (n _TreeNode) String() string {
	var buffer bytes.Buffer
	if n.proc != nil {
		buffer.WriteString("#")
		buffer.WriteString(fmt.Sprint(n.proc.Pid()))
		buffer.WriteString("/")
		buffer.WriteString(n.proc.Executable())
	} else {
		buffer.WriteString("unknown")
	}
	buffer.WriteString("[")

	for i, child := range n.children {
		if i != 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(fmt.Sprint(child.proc.Pid()))
	}
	buffer.WriteString("]")

	return buffer.String()
}

func main() {
	procs, err := ps.Processes()
	if err != nil {
		panic(err)
	}

	nodes := make(map[int]*_TreeNode)

	for _, proc := range procs {
		if proc.Executable() == "ruby" {
			if _, ok := nodes[proc.PPid()]; !ok {
				nodes[proc.PPid()] = &_TreeNode{nil, nil}
			}

			if node, ok := nodes[proc.Pid()]; ok {
				node.proc = proc
			} else {
				nodes[proc.Pid()] = &_TreeNode{proc, nil}
			}

			nodes[proc.PPid()].children = append(nodes[proc.PPid()].children, nodes[proc.Pid()])
		}
	}

	if _, ok := nodes[1]; !ok {
		fmt.Println("No ruby process with PPID 1, exit")
		return
	}

	traverseStack := []*_TreeNode{nodes[1]}
	killStack := []ps.Process(nil)

	for len(traverseStack) > 0 {
		head := traverseStack[len(traverseStack)-1]
		traverseStack = traverseStack[:len(traverseStack)-1]
		traverseStack = append(traverseStack, head.children...)

		killStack = append([]ps.Process{head.proc}, killStack...)
	}

	for _, proc := range killStack {
		if proc == nil {
			continue
		}
		goProc, err := os.FindProcess(proc.Pid())
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = goProc.Kill()
		fmt.Println("SIGKILL", goProc.Pid, err)
	}
}
