package ui

func MoveLeft(cursor, _ int) int {
	if cursor > 0 {
		return cursor - 1
	}
	return cursor
}

func MoveRight(cursor, total, _ int) int {
	if cursor < total-1 {
		return cursor + 1
	}
	return cursor
}

func MoveUp(cursor, columns int) int {
	if cursor >= columns {
		return cursor - columns
	}
	return cursor
}

func MoveDown(cursor, total, columns int) int {
	if cursor+columns < total {
		return cursor + columns
	}
	return cursor
}
