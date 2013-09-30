package game

type Field []BigSquare

type BigSquare []Square

type Square byte

const SIZE = 3


func NewField() Field {
    size := SIZE * SIZE
    field := make([]BigSquare, size)
    for i := 0; i < size; i++ {
        field[i] = make([]Square, size)
    }
    return field
}
