// Bot ingênuo: ponto de partida do trabalho prático.
//
// Este bot implementa o protocolo inteiro e joga de forma simples:
//
//   - guarda os chamados abertos num slice;
//   - a cada turno, cada ambulância livre procura, por varredura linear,
//     o chamado mais próximo e anda na direção dele;
//   - cada ambulância ocupada anda na direção do hospital mais próximo.
//
// Ele não verifica se o chamado ainda pode ser alcançado dentro do prazo,
// não impede que duas ambulâncias sigam para o mesmo chamado e ignora os
// bloqueios da etapa 2. No cenário de estresse, a varredura linear estoura
// o limite de tempo por turno.

package bot

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Chamado struct {
	ID       int
	X, Y     int
	Valor    int
	Prazo    int
	atendido bool
}

type No struct {
	Info []int
	Alt  int
	Esq  *No
	Dir  *No
}

func Insere(a *No, v Chamado, xhospital int, yhospital int) *No {
	if a == nil {
		return &No{Info: []int{v.ID}}
	}

	if distancia(xhospital, yhospital, v.X, v.Y) < distancia(a.Info[0].X, a.Info[0].Y, xhospital, yhospital) {

		// inserir na tabela hash e depois acessar o ID do chamado na árvore AVL

		a.Esq = Insere(a.Esq, v, xhospital, yhospital)
	} else if distancia(xhospital, yhospital, v.X, v.Y) > distancia(a.Info[0].X, a.Info[0].Y, xhospital, yhospital) {
		a.Dir = Insere(a.Dir, v, xhospital, yhospital)
	} else {
		a.Info = append(a.Info, v.ID)
		return a
	}

	atualizaAltura(a)
	return rebalanceia(a)
}

func atualizaAltura(a *No) {
	a.Alt = 1 + maior(altura(a.Esq), altura(a.Dir))
}

func altura(a *No) int {
	if a == nil {
		return -1
	}
	return a.Alt
}

func maior(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func rebalanceia(a *No) *No {
	fb := Fator(a)

	if fb > 1 { // pesada à esquerda
		if Fator(a.Esq) < 0 { // caso esquerda-direita: endireita o filho antes
			a.Esq = RotacaoEsquerda(a.Esq)
		}
		return RotacaoDireita(a) // caso esquerda-esquerda
	}

	if fb < -1 { // pesada à direita
		if Fator(a.Dir) > 0 { // caso direita-esquerda
			a.Dir = RotacaoDireita(a.Dir)
		}
		return RotacaoEsquerda(a) // caso direita-direita
	}

	return a
}

func RotacaoEsquerda(a *No) *No {
	b := a.Dir
	a.Dir = b.Esq
	b.Esq = a

	atualizaAltura(a)
	atualizaAltura(b)
	return b
}

func RotacaoDireita(a *No) *No {
	b := a.Esq
	a.Esq = b.Dir
	b.Dir = a

	atualizaAltura(a)
	atualizaAltura(b)
	return b
}

func Fator(a *No) int {
	if a == nil {
		return 0
	}
	return altura(a.Esq) - altura(a.Dir)
}

type Ambulancia struct {
	ID    int
	X, Y  int
	Livre bool
}

type Hospital struct {
	X, Y int
	No   *No
}

var (
	leitor   *bufio.Scanner
	escritor *bufio.Writer

	hospitais []Hospital
	chamados  []Chamado
)

// lerCampos lê a próxima linha da entrada e a divide em campos.
// Devolve nil quando a entrada termina.
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

// direcao devolve o movimento que aproxima (x, y) de (alvoX, alvoY),
// andando primeiro na horizontal.
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

// lerInicio lê a mensagem de início. Este bot só usa os hospitais; os
// bloqueios são lidos e descartados.
func lerInicio() {
	lerCampos() // INICIO largura altura turnos limite_turno limite_preparo equipe num_equipes
	h := inteiro(lerCampos()[1])
	for i := 0; i < h; i++ {
		c := lerCampos()
		hospitais = append(hospitais, Hospital{inteiro(c[0]), inteiro(c[1]), nil})
	}
	k := inteiro(lerCampos()[1])
	for i := 0; i < k; i++ {
		lerCampos() // id x y
	}
	b := inteiro(lerCampos()[1])
	for i := 0; i < b; i++ {
		lerCampos() // x1 y1 x2 y2
	}
	lerCampos() // FIM

}

// removerPorID procura o chamado no slice e o remove, trocando-o pelo
// último elemento. Custo: proporcional ao número de chamados abertos.
func removerPorID(id int) {
	for i := range chamados {
		if chamados[i].ID == id {
			chamados[i] = chamados[len(chamados)-1]
			chamados = chamados[:len(chamados)-1]
			return
		}
	}
}

// removerExpirados mantém só os chamados com prazo maior ou igual ao turno.
func removerExpirados(turno int) {
	vivos := chamados[:0]
	for _, c := range chamados {
		if c.Prazo >= turno {
			vivos = append(vivos, c)
		}
	}
	chamados = vivos
}

func FolhaMenor(a *No) *No {
	if a == nil {
		return nil
	}

	if a.Esq == nil {
		return a
	}

	return FolhaMenor(a.Esq)
}

func jogarTurno(turno int, ambulancias []Ambulancia, hospitais []Hospital) {
	for _, a := range ambulancias {
		mov := "P"
		hospitalProx := hospitais[0]
		distanciaHospital := distancia(a.X, a.Y, hospitalProx.X, hospitalProx.Y)
		if a.Livre {
			melhor := -1

			for c := range hospitais {
				if distancia(a.X, a.Y, hospitais[c].X, hospitais[c].Y) < distanciaHospital {
					hospitalProx = hospitais[c]
					distanciaHospital = distancia(a.X, a.Y, hospitalProx.X, hospitalProx.Y)
				}
			}
			menordistancia := FolhaMenor(hospitalProx.No)
			if melhor == -1 || menordistancia != nil {
				if menordistancia.Info[0].atendido != true {
					if menordistancia.Info[0].Prazo >= turno+distancia(a.X, a.Y, menordistancia.Info[0].X, menordistancia.Info[0].Y) {
						melhor = menordistancia.Info[0].ID
						menordistancia.Info[0].atendido = true
					}
				}
			}

			if melhor != -1 {
				mov = direcao(a.X, a.Y, menordistancia.Info[0].X, menordistancia.Info[0].Y)
			}
		} else {
			mov = direcao(a.X, a.Y, hospitalProx.X, hospitalProx.Y)
		}
		fmt.Fprintln(escritor, a.ID, mov)
	}
	fmt.Fprintln(escritor, "FIM")
	escritor.Flush()
}

func main() {
	leitor = bufio.NewScanner(os.Stdin)
	leitor.Buffer(make([]byte, 1<<16), 1<<20)
	escritor = bufio.NewWriter(os.Stdout)

	lerInicio()
	fmt.Fprintln(escritor, "PRONTO")
	escritor.Flush()

	for {
		c := lerCampos() // TURNO t
		if c == nil {
			return // fim da partida
		}
		turno := inteiro(c[1])

		n := inteiro(lerCampos()[1]) // CHAMADOS n
		for i := 0; i < n; i++ {
			c := lerCampos()
			chamados = append(chamados, Chamado{inteiro(c[0]), inteiro(c[1]), inteiro(c[2]), inteiro(c[3]), inteiro(c[4]), false})
			for _, h := range hospitais {
				h.No = Insere(h.No, Chamado{inteiro(c[0]), inteiro(c[1]), inteiro(c[2]), inteiro(c[3]), inteiro(c[4]), false}, h.X, h.Y)
			}
		}

		r := inteiro(lerCampos()[1]) // RECOLHIDOS r
		for i := 0; i < r; i++ {
			removerPorID(inteiro(lerCampos()[0]))
		}
		removerExpirados(turno)

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

		jogarTurno(turno, ambulancias, hospitais)
	}
}
