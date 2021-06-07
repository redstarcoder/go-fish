package fish

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Direction is a value representing the direction a ><> is swimming.
type Direction byte

const (
	Right Direction = iota
	Down
	Left
	Up
)

var reader chan byte

// Stack is a type representing a stack in ><>. It holds the stack values in S, as well as a register. The
// register may contain data, but will only be considered filled if filledRegister is also true.
type Stack struct {
	S              []float64
	register       float64
	filledRegister bool
}

// NewStack returns a pointer to a Stack populated with s.
func NewStack(s []float64) *Stack {
	return &Stack{S: s}
}

// Register implements "&".
func (s *Stack) Register() {
	if s.filledRegister {
		s.Push(s.register)
		s.filledRegister = false
	} else {
		s.register = s.Pop()
		s.filledRegister = true
	}
}

// Extend implements ":".
func (s *Stack) Extend() {
	s.Push(s.S[len(s.S)-1])
}

// Reverse implements "r".
func (s *Stack) Reverse() {
	newS := make([]float64, len(s.S))
	for i, ii := 0, len(s.S)-1; ii >= 0; i, ii = i+1, ii-1 {
		newS[i] = s.S[ii]
	}
	s.S = newS
}

// SwapTwo implements "$".
func (s *Stack) SwapTwo() {
	x := s.S[len(s.S)-1]
	s.S[len(s.S)-1] = s.S[len(s.S)-2]
	s.S[len(s.S)-2] = x
}

// SwapThree implements "@": with [1,2,3,4], calling "@" results in [,4,2,3].
func (s *Stack) SwapThree() {
	x := s.S[len(s.S)-1]
	y := s.S[len(s.S)-2]
	s.S[len(s.S)-1] = y
	s.S[len(s.S)-2] = s.S[len(s.S)-3]
	s.S[len(s.S)-3] = x
}

// ShiftRight implements "}".
func (s *Stack) ShiftRight() {
	newS := make([]float64, 1, len(s.S))
	newS[0] = s.Pop()
	s.S = append(newS, s.S...)
}

// ShiftLeft implements "{".
func (s *Stack) ShiftLeft() {
	r := s.S[0]
	s.S = s.S[1:]
	s.Push(r)
}

// Push appends r to the end of the stack.
func (s *Stack) Push(r float64) {
	s.S = append(s.S, float64(r))
}

// Pop removes the value on the end of the stack and returns it.
func (s *Stack) Pop() (r float64) {
	if len(s.S) > 0 {
		r = s.S[len(s.S)-1]
		s.S = s.S[:len(s.S)-1]
	} else {
		panic("Stack is empty!")
	}
	return
}

func longestLineLength(lines []string) (l int) {
	for _, s := range lines {
		if len(s) > l {
			l = len(s)
		}
	}
	return
}

// CodeBox is an object usually created with NewCodeBox. It contains a ><> program complete with a stack,
// and is typically run in steps via CodeBox.Swim.
type CodeBox struct {
	fX, fY        int
	fDir          Direction
	width, height int
	box           [][]byte
	stacks        []*Stack
	p             int // Used to keep track of the current stack
	stringMode    byte
	compMode      bool
}

// NewCodeBox returns a pointer to a new CodeBox. "script" should be a complete ><> script, "stack" should
// be the initial stack, and compatibilityMode should be set if fishinterpreter.com behaviour is needed.
func NewCodeBox(script string, stack []float64, compatibilityMode bool) *CodeBox {
	cB := new(CodeBox)

	script = strings.Replace(script, "\r", "", -1)
	if len(script) == 0 || script == "\n" {
		panic("Cannot accept script of length 0 (No room for the fish to survive).")
	}

	lines := strings.Split(script, "\n")
	cB.width = longestLineLength(lines)
	cB.height = len(lines)

	cB.box = make([][]byte, cB.height)
	for i, s := range lines {
		cB.box[i] = make([]byte, cB.width)
		for ii, r := 0, byte(0); ii < cB.width; ii++ {
			if ii < len(s) {
				r = byte(s[ii])
			} else {
				r = ' '
			}
			cB.box[i][ii] = byte(r)
		}
	}

	cB.stacks = []*Stack{NewStack(stack)}
	cB.compMode = compatibilityMode

	return cB
}

// Exe executes the instruction the ><> is currently on top of. It returns true when it executes ";".
func (cB *CodeBox) Exe(r byte) bool {
	switch r {
	default:
		panic(r)
	case ' ':
	case ';':
		return true
	case '>':
		cB.fDir = Right
	case 'v':
		cB.fDir = Down
	case '<':
		cB.fDir = Left
	case '^':
		cB.fDir = Up
	case '|':
		if cB.fDir == Right {
			cB.fDir = Left
		} else if cB.fDir == Left {
			cB.fDir = Right
		}
	case '_':
		if cB.fDir == Down {
			cB.fDir = Up
		} else if cB.fDir == Up {
			cB.fDir = Down
		}
	case '#':
		switch cB.fDir {
		case Right:
			cB.fDir = Left
		case Down:
			cB.fDir = Up
		case Left:
			cB.fDir = Right
		case Up:
			cB.fDir = Down
		}
	case '/':
		switch cB.fDir {
		case Right:
			cB.fDir = Up
		case Down:
			cB.fDir = Left
		case Left:
			cB.fDir = Down
		case Up:
			cB.fDir = Right
		}
	case '\\':
		switch cB.fDir {
		case Right:
			cB.fDir = Down
		case Down:
			cB.fDir = Right
		case Left:
			cB.fDir = Up
		case Up:
			cB.fDir = Left
		}
	case 'x':
		cB.fDir = Direction(rand.Int31n(4))
	case '"', '\'':
		if cB.stringMode == 0 {
			cB.stringMode = r
		} else if r == cB.stringMode {
			cB.stringMode = 0
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		cB.Push(float64(r - '0'))
	case 'a', 'b', 'c', 'd', 'e', 'f':
		cB.Push(float64(r - 'a' + 10))
	case '&':
		cB.Register()
	case 'o':
		fmt.Print(string(byte(cB.Pop())))
	case 'n':
		fmt.Printf("%v", cB.Pop())
	case 'r':
		cB.ReverseStack()
	case '+':
		cB.Push(cB.Pop() + cB.Pop())
	case '-':
		x := cB.Pop()
		y := cB.Pop()
		cB.Push(y - x)
	case '*':
		cB.Push(cB.Pop() * cB.Pop())
	case ',':
		x := cB.Pop()
		y := cB.Pop()
		cB.Push(y / x)
	case '%':
		x := cB.Pop()
		y := cB.Pop()
		cB.Push(float64(int64(y) % int64(x)))
	case '=':
		if cB.Pop() == cB.Pop() {
			cB.Push(1)
		} else {
			cB.Push(0)
		}
	case ')':
		x := cB.Pop()
		y := cB.Pop()
		if y > x {
			cB.Push(1)
		} else {
			cB.Push(0)
		}
	case '(':
		x := cB.Pop()
		y := cB.Pop()
		if y < x {
			cB.Push(1)
		} else {
			cB.Push(0)
		}
	case '!':
		cB.Move()
	case '?':
		if cB.Pop() == 0 {
			cB.Move()
		}
	case '.':
		cB.fY = int(cB.Pop())
		cB.fX = int(cB.Pop())
	case ':':
		cB.ExtendStack()
	case '~':
		cB.Pop()
	case '$':
		cB.StackSwapTwo()
	case '@':
		cB.StackSwapThree()
	case '}':
		cB.StackShiftRight()
	case '{':
		cB.StackShiftLeft()
	case ']':
		cB.CloseStack()
	case '[':
		cB.NewStack(int(cB.Pop()))
	case 'l':
		cB.Push(cB.StackLength())
	case 'g':
		cB.Push(float64(cB.box[int(cB.Pop())][int(cB.Pop())]))
	case 'p':
		cB.box[int(cB.Pop())][int(cB.Pop())] = byte(cB.Pop())
	case 'i':
		r := float64(-1)
		b := byte(0)
		select {
		case b = <-reader:
			r = float64(b)
		default:
		}
		cB.Push(r)
	}
	return false
}

// Move changes the fish's x/y coordinates based on CodeBox.fDir.
func (cB *CodeBox) Move() {
	switch cB.fDir {
	case Right:
		cB.fX++
		if cB.fX >= cB.width {
			cB.fX = 0
		}
	case Down:
		cB.fY++
		if cB.fY >= cB.height {
			cB.fY = 0
		}
	case Left:
		cB.fX--
		if cB.fX < 0 {
			cB.fX = cB.width - 1
		}
	case Up:
		cB.fY--
		if cB.fY < 0 {
			cB.fY = cB.height - 1
		}
	}
}

// Swim causes the ><> to execute an instruction, then move. It returns true when it encounters ";".
func (cB *CodeBox) Swim() bool {
	defer func() {
		if r := recover(); r != nil {
			cB.PrintBox()
			fmt.Println("Stack:", cB.Stack())
			fmt.Println("something smells fishy...")
			os.Exit(1)
		}
	}()

	if r := cB.box[cB.fY][cB.fX]; cB.stringMode != 0 && r != cB.stringMode {
		cB.Push(float64(r))
	} else if cB.Exe(r) {
		return true
	}
	cB.Move()
	return false
}

// Stack returns the underlying Stack slice.
func (cB *CodeBox) Stack() []float64 {
	return cB.stacks[cB.p].S
}

// Push appends r to the end of the current stack.
func (cB *CodeBox) Push(r float64) {
	cB.stacks[cB.p].Push(r)
}

// Pop removes the value on the end of the current stack and returns it.
func (cB *CodeBox) Pop() float64 {
	return cB.stacks[cB.p].Pop()
}

// StackLength implements "l" on the current stack.
func (cB *CodeBox) StackLength() float64 {
	return float64(len(cB.stacks[cB.p].S))
}

// Register implements "&" on the current stack.
func (cB *CodeBox) Register() {
	cB.stacks[cB.p].Register()
}

// ReverseStack implements "r" on the current stack.
func (cB *CodeBox) ReverseStack() {
	cB.stacks[cB.p].Reverse()
}

// ExtendStack implements ":" on the current stack.
func (cB *CodeBox) ExtendStack() {
	cB.stacks[cB.p].Extend()
}

// StackSwapTwo implements "$" on the current stack.
func (cB *CodeBox) StackSwapTwo() {
	cB.stacks[cB.p].SwapTwo()
}

// StackSwapThree implements "@" on the current stack.
func (cB *CodeBox) StackSwapThree() {
	cB.stacks[cB.p].SwapThree()
}

// StackShiftRight implements "}" on the current stack.
func (cB *CodeBox) StackShiftRight() {
	cB.stacks[cB.p].ShiftRight()
}

// StackShiftLeft implements "{" on the current stack.
func (cB *CodeBox) StackShiftLeft() {
	cB.stacks[cB.p].ShiftLeft()
}

// CloseStack implements "]".
func (cB *CodeBox) CloseStack() {
	cB.p--
	if cB.compMode {
		cB.stacks[cB.p+1].Reverse() // This is done to match the fishlanguage.com interpreter...
	}
	cB.stacks[cB.p].S = append(cB.stacks[cB.p].S, cB.stacks[cB.p+1].S...)
}

// NewStack implements "[".
func (cB *CodeBox) NewStack(n int) {
	cB.p++
	if cB.p == len(cB.stacks) {
		cB.stacks = append(cB.stacks, NewStack(cB.stacks[cB.p-1].S[len(cB.stacks[cB.p-1].S)-n:]))
		cB.stacks[cB.p-1].S = cB.stacks[cB.p-1].S[:len(cB.stacks[cB.p-1].S)-n]
	} else {
		cB.stacks[cB.p].S = cB.stacks[cB.p-1].S[len(cB.stacks[cB.p-1].S)-n:]
		cB.stacks[cB.p].filledRegister = false
	}
	if cB.compMode {
		cB.stacks[cB.p].Reverse() // This is done to match the fishlanguage.com interpreter...
	}
}

// PrintBox outputs the codebox to stdout.
func (cB *CodeBox) PrintBox() {
	fmt.Println()
	for y, line := range cB.box {
		for x, r := range line {
			if x != cB.fX || y != cB.fY {
				fmt.Print(string(rune(r)))
			} else {
				fmt.Print("\u001b[42m" + string(rune(r)) + "\u001b[0m")
			}
		}
		fmt.Println()
	}
}

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
	reader = make(chan byte, 1024)
	go func() {
		var err error
		b := make([]byte, 1024)
		for err == nil {
			n, err := os.Stdin.Read(b)
			if err == nil {
				for i := 0;i < n;i++ {
					reader <- b[i]
				}
			}
		}
	}()
}
