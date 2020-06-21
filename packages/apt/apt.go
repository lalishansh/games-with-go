package apt

import (
	"math"
	"math/rand"
	"reflect"
	"strconv"

	"github.com/t-RED-69/games-with-go/packages/noise"
)

type Node interface {
	Eval(x, y float32) float32
	String() string
	SetParent(node Node)
	SetChildren(node []Node)
	GetParent() Node
	GetChildren() []Node
	AddRandom(node Node)
	AddLeaf(node Node) bool
	NodeCount() int
}
type BaseNode struct {
	Parent   Node
	Children []Node
}

func CopyTree(node Node, parent Node) Node {
	copy := reflect.New(reflect.ValueOf(node).Elem().Type()).Interface().(Node)
	switch n := node.(type) {
	case *OpConst:
		copy.(*OpConst).val = n.val
	}
	copy.SetParent(parent)
	nodeChild := node.GetChildren()
	copyChild := make([]Node, len(nodeChild))
	copy.SetChildren(copyChild)
	for i := range copyChild {
		copyChild[i] = CopyTree(nodeChild[i], copy)
	}
	return copy
}
func ReplaceNode(old Node, new Node) {
	oldParent := old.GetParent()
	if oldParent != nil {
		for i, child := range oldParent.GetChildren() {
			if child == old {
				oldParent.GetChildren()[i] = new
			}
		}
	}
	new.SetParent(oldParent)
}
func (node *BaseNode) Eval(x, y float32) float32 {
	panic("Tried to call Eval() on BaseNode!")
}
func (node *BaseNode) String() string {
	panic("Tried to call String() on BaseNode!")
}
func (node *BaseNode) SetParent(parent Node) {
	node.Parent = parent
}
func (node *BaseNode) SetChildren(children []Node) {
	node.Children = children
}
func (node *BaseNode) GetChildren() []Node {
	return node.Children
}
func (node *BaseNode) GetParent() Node {
	return node.Parent
}
func (node *BaseNode) AddRandom(nodeToAdd Node) {
	i := rand.Intn(len(node.Children))
	//fmt.Println("nodeToAdd:",nodeToAdd.String())
	if node.Children[i] == nil {
		nodeToAdd.SetParent(node)
		node.Children[i] = nodeToAdd
	} else {
		node.Children[i].AddRandom(nodeToAdd)
	}
}
func (node *BaseNode) AddLeaf(leaf Node) bool {
	for i, child := range node.Children {
		if child == nil {
			leaf.SetParent(node)
			node.Children[i] = leaf
			return true
		} else if node.Children[i].AddLeaf(leaf) {
			return true
		}
	}
	return false
}
func (node *BaseNode) NodeCount() int {
	count := 1
	for _, child := range node.Children {
		count += child.NodeCount()
	}
	return count
}
func GetNthNode(node Node, n, count int) (Node, int) {
	if n == count {
		return node, count
	}
	var result Node
	for _, child := range node.GetChildren() {
		count++
		result, count = GetNthNode(child, n, count)
		if result != nil {
			return result, count
		}
	}
	return nil, count
}

func Mutate(node Node) Node {
	r := rand.Intn(23)
	var mutatedNode Node
	if r <= 19 {
		mutatedNode = GetRandomNode()
	} else {
		mutatedNode = GetRandomLeaf()
	}

	if node.GetParent() != nil {
		for i, nd := range node.GetParent().GetChildren() {
			if nd == node {
				node.GetParent().GetChildren()[i] = mutatedNode
			}
		}
	}
	for i, child := range node.GetChildren() {
		if i >= len(mutatedNode.GetChildren()) {
			break
		}
		mutatedNode.GetChildren()[i] = child //be-aware that what we are providing to mutatedNode is copy of child not the orignal child
		child.SetParent(mutatedNode)
	}
	for i, child := range mutatedNode.GetChildren() {
		if child == nil {
			leaf := GetRandomLeaf()
			leaf.SetParent(mutatedNode)
			mutatedNode.GetChildren()[i] = leaf
		}
	}
	mutatedNode.SetParent(node.GetParent())
	return mutatedNode
}

type OpPicture struct {
	BaseNode
}

func NewOpPicture() *OpPicture {
	return &OpPicture{BaseNode{nil, make([]Node, 3)}}
}
func (op *OpPicture) Eval(x, y float32) float32 {
	panic("Eval called on root of picture tree")
}
func (op *OpPicture) String() string {
	return "( Picture\n " + op.Children[0].String() + "\n " + op.Children[1].String() + "\n " + op.Children[2].String() + "\n)\n"
}

type OpLerp struct {
	BaseNode
}

func NewOpLerp() *OpLerp {
	return &OpLerp{BaseNode{nil, make([]Node, 3)}}
}
func (op *OpLerp) Eval(x, y float32) float32 {
	a := op.Children[0].Eval(x, y)
	b := op.Children[1].Eval(x, y)
	pct := op.Children[2].Eval(x, y)
	return a + pct*(b-a)
}
func (op *OpLerp) String() string {
	return "( Lerp " + op.Children[0].String() + " " + op.Children[1].String() + " " + op.Children[2].String() + " )"
}

type OpSin struct {
	BaseNode
}

func NewOpSin() *OpSin {
	return &OpSin{BaseNode{nil, make([]Node, 1)}}
}
func (op *OpSin) Eval(x, y float32) float32 {
	return float32(math.Sin(float64(op.Children[0].Eval(x, y))))
}
func (op *OpSin) String() string {
	return "( Sin " + op.Children[0].String() + " )"
}

type OpCos struct {
	BaseNode
}

func NewOpCos() *OpCos {
	return &OpCos{BaseNode{nil, make([]Node, 1)}}
}
func (op *OpCos) Eval(x, y float32) float32 {
	return float32(math.Cos(float64(op.Children[0].Eval(x, y))))
}
func (op *OpCos) String() string {
	return "( Cos " + op.Children[0].String() + " )"
}

type OpAtan struct {
	BaseNode
}

func NewOpAtan() *OpAtan {
	return &OpAtan{BaseNode{nil, make([]Node, 1)}}
}
func (op *OpAtan) Eval(x, y float32) float32 {
	return float32(math.Atan(float64(op.Children[0].Eval(x, y))))
}
func (op *OpAtan) String() string {
	return "( Atan " + op.Children[0].String() + " )"
}

type OpPlus struct {
	BaseNode
}

func NewOpPlus() *OpPlus {
	return &OpPlus{BaseNode{nil, make([]Node, 2)}}
}
func (op *OpPlus) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) + op.Children[1].Eval(x, y)
}
func (op *OpPlus) String() string {
	return "( + " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpMinus struct {
	BaseNode
}

func NewOpMinus() *OpMinus {
	return &OpMinus{BaseNode{nil, make([]Node, 2)}}
}
func (op *OpMinus) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) - op.Children[1].Eval(x, y)
}
func (op *OpMinus) String() string {
	return "( - " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpMult struct {
	BaseNode
}

func NewOpMult() *OpMult {
	return &OpMult{BaseNode{nil, make([]Node, 2)}}
}
func (op *OpMult) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) * op.Children[1].Eval(x, y)
}
func (op *OpMult) String() string {
	return "( * " + op.Children[0].String() + " " + op.Children[1].String() + ")"
}

type OpDiv struct {
	BaseNode
}

func NewOpDiv() *OpDiv {
	return &OpDiv{BaseNode{nil, make([]Node, 2)}}
}
func (op *OpDiv) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) / op.Children[1].Eval(x, y)
}
func (op *OpDiv) String() string {
	return "( / " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpAtan2 struct {
	BaseNode
}

func NewOpAtan2() *OpAtan2 {
	return &OpAtan2{BaseNode{nil, make([]Node, 2)}}
}
func (op *OpAtan2) Eval(x, y float32) float32 {
	return float32(math.Atan2(float64(op.Children[0].Eval(x, y)), float64(op.Children[1].Eval(x, y))))
}
func (op *OpAtan2) String() string {
	return "( Atan2 " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpX struct {
	BaseNode
}

func NewOpX() *OpX {
	return &OpX{BaseNode{nil, make([]Node, 0)}}
}
func (op *OpX) Eval(x, y float32) float32 {
	return x
}
func (op *OpX) String() string {
	return "X"
}

type OpY struct {
	BaseNode
}

func NewOpY() *OpY {
	return &OpY{BaseNode{nil, make([]Node, 0)}}
}
func (op *OpY) Eval(x, y float32) float32 {
	return y
}
func (op *OpY) String() string {
	return "Y"
}

type OpConst struct {
	BaseNode
	val float32
}

func NewOpConst() *OpConst {
	return &OpConst{BaseNode{nil, make([]Node, 0)}, rand.Float32()*2 - 1}
}
func (op *OpConst) Eval(x, y float32) float32 {
	return op.val
}
func (op *OpConst) String() string {
	return strconv.FormatFloat(float64(op.val), 'f', 9, 32)
}

type OpNoise struct {
	BaseNode
}

func NewOpNoise() *OpNoise {
	return &OpNoise{BaseNode{nil, make([]Node, 2)}}
}
func (op *OpNoise) Eval(x, y float32) float32 {
	return 80*noise.Snoise2(op.Children[0].Eval(x, y), op.Children[1].Eval(x, y)) - 2
}
func (op *OpNoise) String() string {
	return "( SimplexNoise " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpFBM struct {
	BaseNode
}

func NewOpFBM() *OpFBM {
	return &OpFBM{BaseNode{nil, make([]Node, 3)}}
}
func (op *OpFBM) Eval(x, y float32) float32 {
	return 2*3.627*noise.Fbm2(op.Children[0].Eval(x, y), op.Children[1].Eval(x, y), 5*op.Children[2].Eval(x, y), 0.5, 2, 3) + 0.492 - 1
}
func (op *OpFBM) String() string {
	return "( FBM " + op.Children[0].String() + " " + op.Children[1].String() + " " + op.Children[2].String() + " )"
}

type OpTurbulence struct {
	BaseNode
}

func NewOpTurbulence() *OpTurbulence {
	return &OpTurbulence{BaseNode{nil, make([]Node, 3)}}
}
func (op *OpTurbulence) Eval(x, y float32) float32 {
	return 2*6.96*noise.Turbulance(op.Children[0].Eval(x, y), op.Children[1].Eval(x, y), 5*op.Children[2].Eval(x, y), 0.5, 2, 3) - 1
}
func (op *OpTurbulence) String() string {
	return "( Turbulence " + op.Children[0].String() + " " + op.Children[1].String() + " " + op.Children[2].String() + " )"
}

type OpSquare struct {
	BaseNode
}

func NewOpSquare() *OpSquare {
	return &OpSquare{BaseNode{nil, make([]Node, 1)}}
}
func (op *OpSquare) Eval(x, y float32) float32 {
	value := op.Children[0].Eval(x, y)
	return value * value
}
func (op *OpSquare) String() string {
	return "( Square " + op.Children[0].String() + " )"
}

type OpNegate struct {
	BaseNode
}

func NewOpNegate() *OpNegate {
	return &OpNegate{BaseNode{nil, make([]Node, 1)}}
}
func (op *OpNegate) Eval(x, y float32) float32 {
	return -op.Children[0].Eval(x, y)
}
func (op *OpNegate) String() string {
	return "( Negate " + op.Children[0].String() + " )"
}

type OpCeil struct {
	BaseNode
}

func NewOpCeil() *OpCeil {
	return &OpCeil{BaseNode{nil, make([]Node, 1)}}
}
func (op *OpCeil) Eval(x, y float32) float32 {
	return float32(math.Ceil(float64(op.Children[0].Eval(x, y))))
}
func (op *OpCeil) String() string {
	return "( Ceil " + op.Children[0].String() + " )"
}

type OpLog2 struct {
	BaseNode
}

func NewOpLog2() *OpLog2 {
	return &OpLog2{BaseNode{nil, make([]Node, 1)}}
}
func (op *OpLog2) Eval(x, y float32) float32 {
	return float32(math.Log2(float64(op.Children[0].Eval(x, y))))
}
func (op *OpLog2) String() string {
	return "( Log2 " + op.Children[0].String() + " )"
}

type OpAbs struct {
	BaseNode
}

func NewOpAbs() *OpAbs {
	return &OpAbs{BaseNode{nil, make([]Node, 1)}}
}
func (op *OpAbs) Eval(x, y float32) float32 {
	return float32(math.Abs(float64(op.Children[0].Eval(x, y))))
}
func (op *OpAbs) String() string {
	return "( Abs " + op.Children[0].String() + " )"
}

type OpClip struct {
	BaseNode
}

func NewOpClip() *OpClip {
	return &OpClip{BaseNode{nil, make([]Node, 2)}}
}
func (op *OpClip) Eval(x, y float32) float32 {
	value := op.Children[0].Eval(x, y)
	max := float32(math.Abs(float64(op.Children[1].Eval(x, y))))
	if value > max {
		return max
	} else if value < -max {
		return -max
	}
	return value
}
func (op *OpClip) String() string {
	return "( Clip " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpFloor struct {
	BaseNode
}

func NewOpFloor() *OpFloor {
	return &OpFloor{BaseNode{nil, make([]Node, 1)}}
}
func (op *OpFloor) Eval(x, y float32) float32 {
	return float32(math.Floor(float64(op.Children[0].Eval(x, y))))
}
func (op *OpFloor) String() string {
	return "( Floor " + op.Children[0].String() + " )"
}

type OpWrap struct {
	BaseNode
}

func NewOpWrap() *OpWrap {
	return &OpWrap{BaseNode{nil, make([]Node, 1)}}
}
func (op *OpWrap) Eval(x, y float32) float32 {
	f := op.Children[0].Eval(x, y)
	temp := (f + 1) / 2
	return -1 + 2*(temp-float32(math.Floor(float64(temp))))
}
func (op *OpWrap) String() string {
	return "( Wrap " + op.Children[0].String() + " )"
}

func GetRandomNode() Node {
	r := rand.Intn(20)
	switch r {
	case 0:
		return NewOpPlus()
	case 1:
		return NewOpMinus()
	case 2:
		return NewOpMult()
	case 3:
		return NewOpDiv()
	case 4:
		return NewOpSin()
	case 5:
		return NewOpCos()
	case 6:
		return NewOpAtan()
	case 7:
		return NewOpAtan2()
	case 8:
		return NewOpNoise()
	case 9:
		return NewOpSquare()
	case 10:
		return NewOpLog2()
	case 11:
		return NewOpCeil()
	case 12:
		return NewOpFloor()
	case 13:
		return NewOpLerp()
	case 14:
		return NewOpAbs()
	case 15:
		return NewOpClip()
	case 16:
		return NewOpWrap()
	case 17:
		return NewOpFBM()
	case 18:
		return NewOpTurbulence()
	case 19:
		return NewOpNegate()
	}
	panic("Get Random Noise Failed")
}

func GetRandomLeaf() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return NewOpX()
	case 1:
		return NewOpY()
	case 2:
		return NewOpConst()
	}
	panic("Get Random Leaf Failed")
}
