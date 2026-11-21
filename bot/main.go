// Bot de despacho (etapa 1).
//
// Estruturas:
//   - tabela hash id -> Chamado: achar/remover por id em O(1);
//   - AVL por hospital (duas: comuns e graves), ordenada por (distância do
//     chamado ao hospital, id): os chamados mais próximos ficam à esquerda;
//     cada chamado entra só na árvore do hospital mais próximo dele;
//   - baldes por prazo: expirar chamados sem varrer todos;
//   - reserva: o chamado escolhido sai da árvore, então duas ambulâncias
//     nossas nunca seguem para o mesmo chamado.
package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Chamado struct {
	ID, X, Y  int
	Valor     int
	Prazo     int
	Arv       int
	Dist      int
	Reservado bool
}

type Ambulancia struct {
	ID    int
	X, Y  int
	Livre bool
}

type Hospital struct {
	X, Y int
}

type no struct {
	dist, id int32
	esq, dir int32
	alt      int8
}

var (
	nos   = append(make([]no, 0, 1<<20), no{alt: -1})
	livre int32
)

func novoNo(d, id int32) int32 {
	if livre != 0 {
		i := livre
		livre = nos[i].esq
		nos[i] = no{dist: d, id: id}
		return i
	}
	nos = append(nos, no{dist: d, id: id})
	return int32(len(nos) - 1)
}

func liberaNo(i int32) {
	nos[i] = no{esq: livre}
	livre = i
}

func altura(i int32) int { return int(nos[i].alt) }

func atualizaAltura(i int32) {
	nos[i].alt = int8(1 + maior(altura(nos[i].esq), altura(nos[i].dir)))
}

func maior(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func fator(i int32) int {
	if i == 0 {
		return 0
	}
	return altura(nos[i].esq) - altura(nos[i].dir)
}

func menorChave(d1, i1, d2, i2 int32) bool {
	return d1 < d2 || (d1 == d2 && i1 < i2)
}

func rotacaoEsquerda(a int32) int32 {
	b := nos[a].dir
	nos[a].dir = nos[b].esq
	nos[b].esq = a
	atualizaAltura(a)
	atualizaAltura(b)
	return b
}

func rotacaoDireita(a int32) int32 {
	b := nos[a].esq
	nos[a].esq = nos[b].dir
	nos[b].dir = a
	atualizaAltura(a)
	atualizaAltura(b)
	return b
}

func rebalanceia(a int32) int32 {
	fb := fator(a)
	if fb > 1 {
		if fator(nos[a].esq) < 0 {
			e := rotacaoEsquerda(nos[a].esq)
			nos[a].esq = e
		}
		return rotacaoDireita(a)
	}
	if fb < -1 {
		if fator(nos[a].dir) > 0 {
			d := rotacaoDireita(nos[a].dir)
			nos[a].dir = d
		}
		return rotacaoEsquerda(a)
	}
	return a
}

func insere(r int32, d, id int32) int32 {
	if r == 0 {
		return novoNo(d, id)
	}
	if menorChave(d, id, nos[r].dist, nos[r].id) {
		e := insere(nos[r].esq, d, id)
		nos[r].esq = e
	} else {
		x := insere(nos[r].dir, d, id)
		nos[r].dir = x
	}
	atualizaAltura(r)
	return rebalanceia(r)
}

func removeArv(r int32, d, id int32) int32 {
	if r == 0 {
		return 0
	}
	if d == nos[r].dist && id == nos[r].id {
		if nos[r].esq == 0 || nos[r].dir == 0 {
			filho := nos[r].esq
			if filho == 0 {
				filho = nos[r].dir
			}
			liberaNo(r)
			return filho
		}
		s := nos[r].dir
		for nos[s].esq != 0 {
			s = nos[s].esq
		}
		sd, si := nos[s].dist, nos[s].id
		nos[r].dist, nos[r].id = sd, si
		x := removeArv(nos[r].dir, sd, si)
		nos[r].dir = x
	} else if menorChave(d, id, nos[r].dist, nos[r].id) {
		e := removeArv(nos[r].esq, d, id)
		nos[r].esq = e
	} else {
		x := removeArv(nos[r].dir, d, id)
		nos[r].dir = x
	}
	atualizaAltura(r)
	return rebalanceia(r)
}

func percorre(raiz int32, f func(n int32) bool) {
	var pilha [64]int32
	sp := 0
	n := raiz
	for n != 0 || sp > 0 {
		for n != 0 {
			pilha[sp] = n
			sp++
			n = nos[n].esq
		}
		sp--
		n = pilha[sp]
		if !f(n) {
			return
		}
		n = nos[n].dir
	}
}

var (
	leitor   *bufio.Scanner
	escritor *bufio.Writer

	hospitais []Hospital
	tab       *Tabela
	raizes    []int32
	expira    [][]int
	alvo      []int
	novos     []Chamado
)

func lerCampos() []string {
	if !leitor.Scan() {
		return nil
	}
	return strings.Fields(leitor.Text())
}

func inteiro(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic("número inválido na entrada: " + s)
	}
	return v
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func distancia(x1, y1, x2, y2 int) int {
	return abs(x1-x2) + abs(y1-y2)
}

func direcao(x, y, alvoX, alvoY int) string {
	switch {
	case x < alvoX:
		return "L"
	case x > alvoX:
		return "O"
	case y < alvoY:
		return "S"
	case y > alvoY:
		return "N"
	}
	return "P"
}

func hospitalMaisProximo(x, y int) (int, int) {
	melhor, md := 0, distancia(x, y, hospitais[0].X, hospitais[0].Y)
	for i := 1; i < len(hospitais); i++ {
		if d := distancia(x, y, hospitais[i].X, hospitais[i].Y); d < md {
			melhor, md = i, d
		}
	}
	return melhor, md
}

func lerInicio() {
	lerCampos() // INICIO ...
	h := inteiro(lerCampos()[1])
	for i := 0; i < h; i++ {
		c := lerCampos()
		hospitais = append(hospitais, Hospital{inteiro(c[0]), inteiro(c[1])})
	}
	k := inteiro(lerCampos()[1])
	for i := 0; i < k; i++ {
		lerCampos() // id x y
	}
	alvo = make([]int, k)
	b := inteiro(lerCampos()[1])
	for i := 0; i < b; i++ {
		lerCampos() // x1 y1 x2 y2
	}
	lerCampos() // FIM
	raizes = make([]int32, 2*len(hospitais))
	tab = Nova(proximoPrimo(1 << 20))
}

func agenda(prazo, id int) {
	if prazo >= len(expira) {
		expira = append(expira, make([][]int, prazo+1-len(expira))...)
	}
	expira[prazo] = append(expira[prazo], id)
}

func novoChamado(c []string) {
	id, x, y, valor, prazo := inteiro(c[0]), inteiro(c[1]), inteiro(c[2]), inteiro(c[3]), inteiro(c[4])
	h, d := hospitalMaisProximo(x, y)
	arv := 2 * h
	if valor >= 60 { // grave
		arv++
	}
	ch := Chamado{ID: id, X: x, Y: y, Valor: valor, Prazo: prazo, Arv: arv, Dist: d}
	tab.InsereTabela(id, ch)
	novos = append(novos, ch)
	raizes[arv] = insere(raizes[arv], int32(d), int32(id))
	agenda(prazo, id)
}

func remover(id int) {
	c := tab.Ptr(id)
	if c == nil {
		return
	}
	if !c.Reservado {
		raizes[c.Arv] = removeArv(raizes[c.Arv], int32(c.Dist), int32(id))
	}
	tab.Remove(id)
}

func expirar(turno int) {
	p := turno - 1
	if p < 0 || p >= len(expira) {
		return
	}
	for _, id := range expira[p] {
		remover(id)
	}
	expira[p] = nil
}

func liberar(id int) {
	c := tab.Ptr(id)
	if c != nil && c.Reservado {
		c.Reservado = false
		raizes[c.Arv] = insere(raizes[c.Arv], int32(c.Dist), int32(id))
	}
}

func escolher(a Ambulancia, turno int) int {
	melhorID, melhorArv := 0, 0
	var melhorDist int32
	melhorScore := -1.0
	for t := range raizes {
		visitados, validos := 0, 0
		percorre(raizes[t], func(n int32) bool {
			visitados++
			c := tab.Ptr(int(nos[n].id))
			if c != nil {
				da := distancia(a.X, a.Y, c.X, c.Y)
				if c.Prazo >= turno+da-1 { // chega a tempo
					validos++
					s := float64(c.Valor) / float64(da+c.Dist)
					if s > melhorScore {
						melhorScore, melhorID, melhorArv, melhorDist = s, c.ID, t, nos[n].dist
					}
				}
			}
			return visitados < 24 && validos < 6
		})
	}
	if melhorID != 0 {
		raizes[melhorArv] = removeArv(raizes[melhorArv], melhorDist, int32(melhorID))
		tab.Ptr(melhorID).Reservado = true
	}
	return melhorID
}

func trocarAlvo(a Ambulancia, turno int) {
	cur := tab.Ptr(alvo[a.ID])
	scoreCur := float64(cur.Valor) / float64(distancia(a.X, a.Y, cur.X, cur.Y)+cur.Dist)
	melhor, melhorScore := -1, scoreCur*1.25
	for i := range novos {
		c := &novos[i]
		da := distancia(a.X, a.Y, c.X, c.Y)
		if c.Prazo < turno+da-1 {
			continue
		}
		if s := float64(c.Valor) / float64(da+c.Dist); s > melhorScore {
			melhor, melhorScore = i, s
		}
	}
	if melhor < 0 {
		return
	}
	n := tab.Ptr(novos[melhor].ID)
	if n == nil || n.Reservado {
		return
	}
	liberar(alvo[a.ID])
	n = tab.Ptr(novos[melhor].ID)
	raizes[n.Arv] = removeArv(raizes[n.Arv], int32(n.Dist), int32(n.ID))
	n.Reservado = true
	alvo[a.ID] = n.ID
}

func jogarTurno(turno int, ambulancias []Ambulancia) {
	for _, a := range ambulancias {
		mov := "P"
		if !a.Livre {
			alvo[a.ID] = 0
			h, _ := hospitalMaisProximo(a.X, a.Y)
			mov = direcao(a.X, a.Y, hospitais[h].X, hospitais[h].Y)
		} else {
			if id := alvo[a.ID]; id != 0 {
				c := tab.Ptr(id)
				if c == nil {
					alvo[a.ID] = 0
				} else if c.Prazo < turno+distancia(a.X, a.Y, c.X, c.Y)-1 {
					liberar(id)
					alvo[a.ID] = 0
				}
			}
			if alvo[a.ID] != 0 {
				trocarAlvo(a, turno)
			} else {
				alvo[a.ID] = escolher(a, turno)
			}
			if id := alvo[a.ID]; id != 0 {
				c := tab.Ptr(id)
				mov = direcao(a.X, a.Y, c.X, c.Y)
			}
		}
		escritor.WriteString(strconv.Itoa(a.ID))
		escritor.WriteString(" ")
		escritor.WriteString(mov)
		escritor.WriteString("\n")
	}
	escritor.WriteString("FIM\n")
	escritor.Flush()
}

func main() {
	leitor = bufio.NewScanner(os.Stdin)
	leitor.Buffer(make([]byte, 1<<16), 1<<20)
	escritor = bufio.NewWriter(os.Stdout)

	lerInicio()
	escritor.WriteString("PRONTO\n")
	escritor.Flush()

	for {
		c := lerCampos()
		if c == nil {
			return
		}
		turno := inteiro(c[1])

		novos = novos[:0]
		n := inteiro(lerCampos()[1])
		for i := 0; i < n; i++ {
			novoChamado(lerCampos())
		}

		r := inteiro(lerCampos()[1])
		for i := 0; i < r; i++ {
			remover(inteiro(lerCampos()[0]))
		}
		expirar(turno)

		k := inteiro(lerCampos()[1]) // AMBULANCIAS k
		ambulancias := make([]Ambulancia, k)
		for i := 0; i < k; i++ {
			c := lerCampos()
			ambulancias[i] = Ambulancia{inteiro(c[0]), inteiro(c[1]), inteiro(c[2]), c[3] == "LIVRE"}
		}

		e := inteiro(lerCampos()[1]) // PLACAR e
		for i := 0; i < e; i++ {
			lerCampos()
		}
		lerCampos() // FIM

		jogarTurno(turno, ambulancias)
	}
}
