package jsonnocomment

import (
	"os"
)

// ReadFileAndCleanComment - Функция чтения JSON файла и удаления из него закомментированных строк и блок
// (comments style like С/С++)
func ReadFileAndCleanComment(pathFile string) (cleanFile []byte, err error) {
	file, err := os.ReadFile(pathFile)
	if err != nil {
		return nil, err
	}

	l := len(file)
	cleanFile = make([]byte, l) // Создаем байтовый массив размера прочитанного, чтобы избежать перевыделения памяти
	var (
		commentLine  bool // Признак комментариев (закомментирована вся линия // ТУТ КОМЕНТ ДО КОНЦА СТРОЧКИ\n)
		commentBlock bool // Признак комментариев (закомментирован только блок /* КОММЕНТ ТОЛЬКО ТУТ*/)
		j            int  // Текущий индекс для заполнения среза cleanFile
	)

	for i := 0; i < l; i++ {
		c := file[i]
		switch c {
		case 10: // ASCI CODE OF '/n'
			commentLine = false
			cleanFile[j] = c
			j++
		case 47: // ASCII CODE OF '/'
			if !commentLine && !commentBlock {
				i++
				if i < l {
					switch file[i] { // next symbol
					case 47: // проверка двух скобок подряд '//'  Comment Line Start
						commentLine = true
					case 42: // проверка '/*'   Comment Block Start
						commentBlock = true
					default:
						i--
						cleanFile[j] = c
						j++
					}
				} else {
					i--
				}
			}
		case 42: // ASCI CODE OF '*'
			if !commentLine {
				i++
				if i < l {
					switch file[i] { // next symbol
					case 47: // проверка на '*/'  Comment Block End
						if commentBlock {
							commentBlock = false
						}
					default:
						i--
						if !commentBlock {
							cleanFile[j] = c
							j++
						}
					}
				} else {
					i--
				}
			}
		default: // any other symbols
			if !commentLine && !commentBlock { // not comment line
				cleanFile[j] = c
				j++
			}
		}
	}
	return cleanFile[:j], err
}
