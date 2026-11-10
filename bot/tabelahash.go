package bot

import (
	"fmt"
	"strings"
)

type Tabela struct {
	chaves  []int
	valores []int
	ocupada []bool
	n       int // número de pares guardados
}

func Nova(m int) *Tabela {
	return &Tabela{
		chaves:  make([]int, m),
		valores: make([]int, m),
		ocupada: make([]bool, m),
	}
}

func (t *Tabela) Len() int {
	return t.n
}

func (t *Tabela) m() int {
	return len(t.chaves)
}

const misturar = false

func (t *Tabela) hash(k int) int {
	if misturar {
		k = mistura(k)
	}
	h := k % t.m()
	if h < 0 {
		h += t.m()
	}
	return h
}

func mistura(k int) int {
	x := uint64(k)
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return int(x >> 1) // descarta o bit de sinal, para o resultado ser positivo
}

func (t *Tabela) distancia(de, ate int) int {
	return (ate - de + t.m()) % t.m()
}

func (t *Tabela) Busca(k int) (int, bool) {
	p := t.hash(k)
	for t.ocupada[p] {
		if t.chaves[p] == k {
			return t.valores[p], true
		}
		p = (p + 1) % t.m()
	}
	return 0, false
}

func (t *Tabela) String() string {
	var b strings.Builder
	for p := 0; p < t.m(); p++ {
		fmt.Fprintf(&b, "%4d", p)
	}
	b.WriteString("\n")
	for p := 0; p < t.m(); p++ {
		if t.ocupada[p] {
			fmt.Fprintf(&b, "%4d", t.chaves[p])
		} else {
			b.WriteString("   .")
		}
	}
	return b.String()
}

func proximoPrimo(k int) int {
	for {
		primo := k >= 2
		for d := 2; d*d <= k; d++ {
			if k%d == 0 {
				primo = false
				break
			}
		}
		if primo {
			return k
		}
		k++
	}
}

func (t *Tabela) InsereTabela(k, v int) {
	p := t.hash(k)
	for t.ocupada[p] {
		if t.chaves[p] == k {
			t.valores[p] = v
			return
		}
		p = (p + 1) % t.m()
	}
	t.chaves[p] = k
	t.valores[p] = v
	t.ocupada[p] = true
	t.n++
	if 10*t.n > 7*t.m() {
		t.redimensiona(proximoPrimo(2 * t.m()))
	}
}

func (t *Tabela) redimensiona(m int) {
	tabela := Nova(m)
	for i, val := range t.valores {
		if t.ocupada[i] {
			tabela.InsereTabela(t.chaves[i], val)
		}
	}
	*t = *tabela

}

func (t *Tabela) Remove(chave int) bool {

	for i := t.hash(chave); t.ocupada[i]; i = (i + 1) % t.m() {
		if t.ocupada[i] {
			if t.chaves[i] == chave {
				t.ocupada[i] = false
				t.n--
				for j := (i + 1) % t.m(); t.ocupada[j]; j = (j + 1) % t.m() {
					h := t.hash(t.chaves[j])
					if t.distancia(h, i) < t.distancia(h, j) {
						t.chaves[i] = t.chaves[j]
						t.valores[i] = t.valores[j]
						t.ocupada[i] = true
						t.ocupada[j] = false
						i = j
					}
				}
				return true
			}

		}
	}
	return false
}
