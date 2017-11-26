package main

import (
    "fmt"
    "errors"
    "unicode"
    "bytes"
    "strings"
)

const (
    ERR_OOB = "You can't move your piece out of bounds"
    ERR_COLL = "You can't jump if you're not a knight"
    ERR_NOMOVE = "You have to move somewhere"
    ERR_MOVEPP = "You can't move that much"
    ERR_TK = "You can't take your own pieces"
    ERR_AXIS = "You can't move like that"
    ERR_NO_SUCH_PIECE = "There's no piece there"
    ERR_SYNTAX = "Move syntax error"
)

var g GameBoard

type GameBoard struct {
    Ok bool
    Locations [][]rune
    White []Piece
    Black []Piece
    
}

func (g *GameBoard) Print() string {
    var buffer bytes.Buffer
    
    buffer.WriteString(fmt.Sprintf("%2c%2c%2c%2c%2c%2c%2c%2c%2c%2c\n", ' ', 'a','b','c','d','e','f','g','h',' '))
    
    for i := 7; i >= 0; i-- {
        buffer.WriteString(fmt.Sprintf("%2d", i+1))
        for j := 0; j < 8; j++ {
            buffer.WriteString(fmt.Sprintf("%2c", g.Locations[j][i]))
        }
        buffer.WriteString(fmt.Sprintf("%2d\n", i+1))
        
    }
    
    buffer.WriteString(fmt.Sprintf("%2c%2c%2c%2c%2c%2c%2c%2c%2c%2c\n", ' ', 'a','b','c','d','e','f','g','h',' '))
    
    return buffer.String()
    
}

func (g *GameBoard) Pprint() string {
    var buffer bytes.Buffer
    
    sym := map[rune]rune {
        'k' : '♔',
        'q' : '♕',
        'r' : '♖',
        'b' : '♗',
        'n' : '♘',
        'p' : '♙',
        'K' : '♚',
        'Q' : '♛',
        'R' : '♜',
        'B' : '♝',
        'N' : '♞',
        'P' : '♟',
        ' ' : ' ',
    }
    
    buffer.WriteString(fmt.Sprintf("%4c%4c%4c%4c%4c%4c%4c%4c%4c%4c\n", ' ', 'a','b','c','d','e','f','g','h',' '))
    
    for i := 7; i >= 0; i-- {
        buffer.WriteString(fmt.Sprintf("%4d", i+1))
        for j := 0; j < 8; j++ {
            buffer.WriteString(fmt.Sprintf("%4c", sym[g.Locations[j][i]]))
        }
        buffer.WriteString(fmt.Sprintf("%4d\n", i+1))
        
    }
    
    buffer.WriteString(fmt.Sprintf("%4c%4c%4c%4c%4c%4c%4c%4c%4c%4c\n", ' ', 'a','b','c','d','e','f','g','h',' '))
    
    return buffer.String()
    
}

func (g *GameBoard) Take(p Point) error {
    piece, err := g.GetPiece(p.X, p.Y)
    if err != nil {
        return errors.New(ERR_NO_SUCH_PIECE)
    }
    
    piece.Get().Taken = true
    return nil
}

func (g *GameBoard) GetPiece(i, j int) (p Piece, err error) {
    r := g.Locations[i][j]
    
    if r == ' ' {
        err = errors.New(ERR_NO_SUCH_PIECE)
        return p, err
    }
    
    if unicode.IsUpper(r) {
        // search black
        for _, piece := range(g.Black) {
            p := piece.Get()
            if !p.Taken && p.X == i && p.Y == j {
                return piece, nil
            }
        }
    } else {
        for _, piece := range(g.White) {
            p := piece.Get()
            if !p.Taken && p.X == i && p.Y == j {
                return piece, nil
            }
        }
    }
    
    return p, errors.New(ERR_NO_SUCH_PIECE)
}

func (g *GameBoard) Init() {

    r1 := Rook{Point{0,0,'r', false}, g}
    r2 := Rook{Point{7,0,'r', false}, g}
    R1 := Rook{Point{0,7,'R', false}, g}
    R2 := Rook{Point{7,7,'R', false}, g}
    
    n1 := Knight{Point{1,0,'n', false}, g}
    n2 := Knight{Point{6,0,'n', false}, g}
    N1 := Knight{Point{1,7,'N', false}, g}
    N2 := Knight{Point{6,7,'N', false}, g}
    
    b1 := Bishop{Point{2,0,'b', false}, g}
    b2 := Bishop{Point{5,0,'b', false}, g}
    B1 := Bishop{Point{2,7,'B', false}, g}
    B2 := Bishop{Point{5,7,'B', false}, g}
    
    p1 := Pawn{Point{0,1,'p', false}, g}
    p2 := Pawn{Point{1,1,'p', false}, g}
    p3 := Pawn{Point{2,1,'p', false}, g}
    p4 := Pawn{Point{3,1,'p', false}, g}
    p5 := Pawn{Point{4,1,'p', false}, g}
    p6 := Pawn{Point{5,1,'p', false}, g}
    p7 := Pawn{Point{6,1,'p', false}, g}
    p8 := Pawn{Point{7,1,'p', false}, g}
    
    P1 := Pawn{Point{0,6,'P', false}, g}
    P2 := Pawn{Point{1,6,'P', false}, g}
    P3 := Pawn{Point{2,6,'P', false}, g}
    P4 := Pawn{Point{3,6,'P', false}, g}
    P5 := Pawn{Point{4,6,'P', false}, g}
    P6 := Pawn{Point{5,6,'P', false}, g}
    P7 := Pawn{Point{6,6,'P', false}, g}
    P8 := Pawn{Point{7,6,'P', false}, g}
    
    
    k := King{Point{3,0,'k', false}, g}
    K := King{Point{3,7,'K', false}, g}
    
    q := Queen{Point{4,0,'q', false}, g}
    Q := Queen{Point{4,7,'Q', false}, g}
    
    g.White = []Piece{&r1, &r2, &n1, &n2, &b1, &b2, &k, &q, &p1, &p2, &p3, &p4, &p5, &p6, &p7, &p8}
    g.Black = []Piece{&R1, &R2, &N1, &N2, &B1, &B2, &K, &Q, &P1, &P2, &P3, &P4, &P5, &P6, &P7, &P8}
    
    g.Locations = make([][]rune, 8)
    for i := 0; i < 8; i++ {
        g.Locations[i] = make([]rune, 8)
        for j := 0; j < 8; j++ {
            g.Locations[i][j] = ' '
        }
    }
    
    for _, w := range(g.White) {
        p := w.Get()
        g.Locations[p.X][p.Y] = p.Code
    }
    
    for _, b := range(g.Black) {
        p := b.Get()
        g.Locations[p.X][p.Y] = p.Code
    }
    
    g.Ok = true

}

type Point struct {
    X int
    Y int
    Code rune
    Taken bool
}

func (p *Point) Get() *Point {
    return p
}

type Piece interface {
    Move(to Point) error
    Get() *Point
}

type King struct {
    Point
    GB *GameBoard
}

func (p *King) Move(to Point) error {
    if isOoB(to) {
        return errors.New(ERR_OOB)
    }
    
    diffx, diffy := abs(to.X-p.X), abs(to.Y-p.Y)
    
    if diffx == 0 && diffy == 0 {
        return errors.New(ERR_NOMOVE)
    }
    
    if diffx > 1 || diffy > 1 {
        return errors.New(ERR_MOVEPP)
    }
    
    spot := p.GB.Locations[to.X][to.Y]
    if spot != ' ' && !(unicode.IsUpper(spot) && unicode.IsUpper(p.Code)) {
        return errors.New(ERR_TK)
    }
    
    p.GB.Locations[p.X][p.Y] = ' '
    p.GB.Locations[to.X][to.Y] = p.Code
    p.X = to.X
    p.Y = to.Y    
    
    return nil
}

type Queen struct {
    Point
    GB *GameBoard
}

func (p *Queen) Move(to Point) error {
    if isOoB(to) {
        return errors.New(ERR_OOB)
    }
    
    diffx, diffy := abs(to.X-p.X), abs(to.Y-p.Y)
    if diffx == 0 && diffy == 0 {
        return errors.New(ERR_NOMOVE)
    }
        
    var collision bool
    if diffx != 0 && diffy != 0 {
        if diffx-diffy != 0 {
            return errors.New(ERR_AXIS)
        } else {
            collision = checkDCol(p.Point, to, p.GB)
        }
    } else {
        collision = checkHVCol(p.Point, to, p.GB)
    }
    
    if collision {
        return errors.New(ERR_COLL)
    }
    
    spot := p.GB.Locations[to.X][to.Y]
    if spot != ' ' && !(unicode.IsUpper(spot) != unicode.IsUpper(p.Code)) {
        return errors.New(ERR_TK)
    }
    
    p.GB.Take(to)
    p.GB.Locations[p.X][p.Y] = ' '
    p.GB.Locations[to.X][to.Y] = p.Code
    p.X = to.X
    p.Y = to.Y
    return nil
}

type Rook struct {
    Point
    GB *GameBoard
}

func (p *Rook) Move(to Point) error {

    if isOoB(to) {
        return errors.New(ERR_OOB)
    }
    
    diffx, diffy := abs(to.X-p.X), abs(to.Y-p.Y)
    if diffx == 0 && diffy == 0 {
        return errors.New(ERR_NOMOVE)
    }
        
    var collision bool
    if diffx != 0 && diffy != 0 {
        return errors.New(ERR_AXIS)
    } else {
        collision = checkHVCol(p.Point, to, p.GB)
    }    
    if collision {
        return errors.New(ERR_COLL)
    }
    
    spot := p.GB.Locations[to.X][to.Y]
    if spot != ' ' && !(unicode.IsUpper(spot) != unicode.IsUpper(p.Code)) {
        return errors.New(ERR_TK)
    }
    
    p.GB.Take(to)
    p.GB.Locations[p.X][p.Y] = ' '
    p.GB.Locations[to.X][to.Y] = p.Code
    p.X = to.X
    p.Y = to.Y    
    return nil

}


type Bishop struct {
    Point
    GB *GameBoard
}

func (p *Bishop) Move(to Point) error {

    if isOoB(to) {
        return errors.New(ERR_OOB)
    }
    
    diffx, diffy := abs(to.X-p.X), abs(to.Y-p.Y)
    if diffx == 0 && diffy == 0 {
        return errors.New(ERR_NOMOVE)
    }
        
    var collision bool
    if diffx != diffy {
        return errors.New(ERR_AXIS)
    } else {
        collision = checkDCol(p.Point, to, p.GB)
    }    
    if collision {
        return errors.New(ERR_COLL)
    }
    
    spot := p.GB.Locations[to.X][to.Y]
    if spot != ' ' && !(unicode.IsUpper(spot) != unicode.IsUpper(p.Code)) {
        return errors.New(ERR_TK)
    }
    
    p.GB.Take(to)
    p.GB.Locations[p.X][p.Y] = ' '
    p.GB.Locations[to.X][to.Y] = p.Code
    p.X = to.X
    p.Y = to.Y   
    return nil

}


type Knight struct {
    Point
    GB *GameBoard
}

func (p *Knight) Move(to Point) error {

    if isOoB(to) {
        return errors.New(ERR_OOB)
    }
    
    diffx, diffy := abs(to.X-p.X), abs(to.Y-p.Y)
    if diffx == 0 && diffy == 0 {
        return errors.New(ERR_NOMOVE)
    }
        
    if !(diffx == 1 && diffy == 2 || diffx == 2 && diffy == 1) {
        return errors.New(ERR_AXIS)
    }
    
    spot := p.GB.Locations[to.X][to.Y]
    if spot != ' ' && !(unicode.IsUpper(spot) != unicode.IsUpper(p.Code)) {
        return errors.New(ERR_TK)
    }
    
    p.GB.Take(to)
    p.GB.Locations[p.X][p.Y] = ' '
    p.GB.Locations[to.X][to.Y] = p.Code
    p.X = to.X
    p.Y = to.Y  
    return nil

}

type Pawn struct {
    Point
    GB *GameBoard
}

func (p *Pawn) Move(to Point) error {

    if isOoB(to) {
        return errors.New(ERR_OOB)
    }
    
    diffx, diffy := to.X-p.X, to.Y-p.Y
    if diffx == 0 && diffy == 0 {
        return errors.New(ERR_NOMOVE)
    }
    
    if unicode.IsUpper(p.Code) {
        diffx, diffy = -diffx, -diffy
    }
    
    if diffy < 0 || diffx < 0 { // can't go backwards
        return errors.New(ERR_AXIS)
    }    
    if diffx == 0 {
        if diffy == 2 {
            if !(p.Y == 1 || p.Y == 6) {
                return errors.New(ERR_AXIS)
            }
            if checkHVCol(p.Point, to, p.GB) {
                return errors.New(ERR_COLL)
            }
        } else if diffy == 1 {
            // ok
        } else {
            return errors.New(ERR_MOVEPP)
        }
    } else if diffx == 1 {
        if diffy != 1 {
            return errors.New(ERR_AXIS)
        }
        e := p.GB.Locations[to.X][to.Y]
        if e == ' ' {
            return errors.New(ERR_AXIS)
        } else if !(unicode.IsUpper(e) != unicode.IsUpper(p.Code)) {
            return errors.New(ERR_TK)
        } 
    }
    
    p.GB.Take(to)
    p.GB.Locations[p.X][p.Y] = ' '
    p.GB.Locations[to.X][to.Y] = p.Code
    p.X = to.X
    p.Y = to.Y  
    return nil

}


func checkHVCol(from, to Point, g *GameBoard) bool {
    if from.X - to.X == 0 {
        if to.Y > from.Y {
            for i := from.Y+1; i < to.Y; i++ {
                if g.Locations[to.X][i] != ' ' {
                    return true
                }
            }
        } else {
            for i := from.Y-1; i > to.Y; i-- {
                if g.Locations[to.X][i] != ' ' {
                    return true
                }
            }
        }
    } else {
        if to.X > from.X {
            for i := from.X+1; i < to.X; i++ {
                if g.Locations[i][to.Y] != ' ' {
                    return true
                }
            }
        } else {
            for i := from.X-1; i > to.X; i-- {
                if g.Locations[i][to.Y] != ' ' {
                    fmt.Println(i)
                    return true
                }
            }
        }    
    }
    return false
}

func checkDCol(from, to Point, g *GameBoard) bool {
    if to.X > from.X {
        if to.Y > from.Y {
            // UR
            for i, j := from.X+1, from.Y+1; i < to.X; i,j = i+1, j+1 {
                if g.Locations[i][j] != ' ' {
                    return true
                }
            }
        } else {
            // DR
            for i, j := from.X+1, from.Y-1; i < to.X; i, j =i+1, j-1 {
                if g.Locations[to.X][i] != ' ' {
                    return true
                }
            }
        }
    } else {
        if to.Y > from.Y {
            // UL
            for i, j := from.X-1, from.Y+1; i > to.X; i,j=i-1, j+1 { 
                if g.Locations[i][j] != ' ' {
                    return true
                }
            }
        } else {
            // DL
            for i, j := from.X-1, from.Y-1; i > to.X; i,j=i-1, j-1 {
                if g.Locations[i][to.Y] != ' ' {
                    return true
                }
            }
        }    
    }
    return false
}

func isOoB(p Point) bool {
    return p.X < 0 || p.X > 7 || p.Y < 0 || p.X > 7
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}

func parse(s string, g *GameBoard) error {
    if len(s) != 4 {
        return errors.New(ERR_SYNTAX)
    }
    
    conv := map[byte]int {
        'a' : 0,
        'b' : 1,
        'c' : 2,
        'd' : 3,
        'e' : 4,
        'f' : 5,
        'g' : 6,
        'h' : 7,
        '1' : 0,
        '2' : 1,
        '3' : 2,
        '4' : 3,
        '5' : 4,
        '6' : 5,
        '7' : 6,
        '8' : 7}
    
    fx, ok1 := conv[s[0]]
    fy, ok2 := conv[s[1]]
    tx, ok3 := conv[s[2]]
    ty, ok4 := conv[s[3]]
    
    if !(ok1 && ok2 && ok3 && ok4) {
        return errors.New(ERR_SYNTAX)
    }
    
    piece, err := g.GetPiece(fx, fy)
    if err != nil {
        return err
    }
    return piece.Move(Point{tx, ty, ' ', false})   
    
}

func chess(m Message, db *Database) string {
    parts := strings.Fields(m.Text) 
    if len(parts) == 3 {
        if parts[2] == "reset" {
            g.Init()
            return "\n```" + g.Print() + "```"
        } else if parts[2] == "print" {
            if !g.Ok {
                g.Init()
            }
            return "\n```" + g.Print() + "```"
        } else {
            if !g.Ok {
                g.Init()
            }
            err := parse(parts[2], &g)
            if err != nil {
                return err.Error()
            }
            return "\n```" + g.Print() + "```"
        }
    }
    
    return "Idk what you mean"
}