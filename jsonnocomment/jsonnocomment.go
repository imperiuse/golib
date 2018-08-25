package jsonnocomment

import (
	"io/ioutil"
)

// ReadFileAndCleanComment - Функция чтения JSON файла и удаления из него закоментированных строк и блок (комментарии согласно правилам С/С++)
func ReadFileAndCleanComment(pathFile string) (cleanFile []byte, err error) {
	// Чтение из file по pathFile
	file, err := ioutil.ReadFile(pathFile)
	if err != nil {
		return nil, err
	}
	l := len(file)
	// Создаем байтовый массив размера прочитанного чтобы избежать перевыделения памяти
	cleanFile = make([]byte, l)
	var commentLine bool  // признак комментариев (закоментирована вся линия  // ТУТ КОМЕНТ ДО КОНЦА СТРОЧКИ\n)
	var commentBlock bool // признак комментариев (закоментирован только блок  /* КОММЕНТ ТОЛЬКО ТУТ*/)
	var j int             // текущий индекс для заполнения среза cleanFile
	for i := 0; i < l; i++ {
		c := file[i]
		switch c {
		case 10: // ASCI CODE OF '/n'
			commentLine = false
		case 47: // ASCII CODE OF '/'
			i++
			if (i) < l {
				switch file[i] { // next symbol
				case 47: // проверка двух скобок подряд '//'  Comment Line Start
					commentLine = true
				case 42: // проверка '/*'   Comment Block Start
					commentBlock = true
				default:
					i--
				}
			} else {
				i--
			}
		case 42: // ASCI CODE OF '*'
			if commentBlock {
				i++
				if (i) < l {
					switch file[i] { // next symbol
					case 47: // проверка двух скобок подряд '*/'  Comment Block End
						commentBlock = false
					default:
						i--
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
	//fmt.Println(string(cleanFile[:j]))
	return cleanFile[:j], err
}
