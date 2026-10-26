package crossrefapi_test

import (
	"fmt"
	"os"

	"github.com/caltechlibrary/crossrefapi"
)

func ExamplePlainText() {
	// Publishers indent their markup; the newline and spaces are layout, not text.
	title, warnings := crossrefapi.PlainText("<i>Euclid</i>\n                    preparation")
	fmt.Println(title, len(warnings))
	// Output: Euclid preparation 0
}

func ExamplePlainText_scripts() {
	title, _ := crossrefapi.PlainText("Irradiation of H\n  <sub>2</sub>\n  O:NH\n  <sub>3</sub>\n  :CO ices")
	fmt.Println(title)
	// Output: Irradiation of H₂O:NH₃:CO ices
}

func ExamplePlainText_formula() {
	const mathML = `a loop model with<mml:math xmlns:mml="http://www.w3.org/1998/Math/MathML">` +
		`<mml:mn>1</mml:mn><mml:mo>/</mml:mo><mml:msup><mml:mi>r</mml:mi><mml:mn>2</mml:mn></mml:msup>` +
		`</mml:math>current-current interactions`
	title, _ := crossrefapi.PlainText(mathML)
	fmt.Println(title)
	// Output: a loop model with 1/r² current-current interactions
}

func ExamplePlainText_warning() {
	// Unicode has no subscript N, so the formula is written in a linear form
	// and a warning says so.
	const mathML = `with <mml:math xmlns:mml="http://www.w3.org/1998/Math/MathML">` +
		`<mml:msub><mml:mi>Z</mml:mi><mml:mi>N</mml:mi></mml:msub></mml:math> symmetry`
	title, warnings := crossrefapi.PlainText(mathML)
	fmt.Println(title)
	for _, w := range warnings {
		fmt.Println(w)
	}
	// Output:
	// with Z_N symmetry
	// math-linear: Z_N
}

func ExampleHTMLAbstract() {
	html, warnings := crossrefapi.HTMLAbstract(
		"<jats:title>Abstract</jats:title>\n  <jats:p>Water is H<jats:sub>2</jats:sub>O.</jats:p>")
	fmt.Println(html, len(warnings))
	// Output: <p>Water is H<sub>2</sub>O.</p> 0
}

func ExampleHTMLAbstract_formula() {
	const abstract = `<jats:p>The <jats:italic>Euclid</jats:italic> mission, with ` +
		`<mml:math xmlns:mml="http://www.w3.org/1998/Math/MathML"><mml:mi>J</mml:mi><mml:mo>/</mml:mo>` +
		`<mml:mi>ψ</mml:mi></mml:math> data.</jats:p>`
	html, _ := crossrefapi.HTMLAbstract(abstract)
	fmt.Println(html)
	// Output: <p>The <i>Euclid</i> mission, with \(J/\psi\) data.</p>
}

func ExampleMessage_FullTitle() {
	msg := &crossrefapi.Message{
		Title:    []string{"<i>Euclid</i>\n   preparation"},
		SubTitle: []string{"CII. Non-Gaussianity of two-point statistics likelihood"},
	}
	title, _ := msg.FullTitle()
	fmt.Println(title)
	// Output: Euclid preparation: CII. Non-Gaussianity of two-point statistics likelihood
}

func ExampleMessage_AlternateTitles() {
	msg := &crossrefapi.Message{
		Title: []string{"X-chromosomale juvenile Retinoschisis", "X-linked juvenile retinoschisis"},
	}
	// FullTitle uses the first title and reports the others...
	title, warnings := msg.FullTitle()
	fmt.Println(title)
	fmt.Println(warnings[0])
	// ...and AlternateTitles returns them.
	others, _ := msg.AlternateTitles()
	fmt.Println(others)
	// Output:
	// X-chromosomale juvenile Retinoschisis
	// multiple-titles: X-linked juvenile retinoschisis
	// [X-linked juvenile retinoschisis]
}

func ExampleNewNormalizer() {
	// An override file adds to or replaces entries in the default table. This one
	// stops the "Abstract" heading from being dropped.
	f, err := os.CreateTemp("", "symbols-*.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.Remove(f.Name())
	f.WriteString(`{"dropped_headings": []}`)
	f.Close()

	n, err := crossrefapi.NewNormalizer(f.Name())
	if err != nil {
		fmt.Println(err)
		return
	}
	const abstract = "<jats:title>Abstract</jats:title><jats:p>Text.</jats:p>"
	withOverride, _ := n.HTMLAbstract(abstract)
	withDefault, _ := crossrefapi.HTMLAbstract(abstract)
	fmt.Println(withOverride)
	fmt.Println(withDefault)
	// Output:
	// <p><b>Abstract</b></p><p>Text.</p>
	// <p>Text.</p>
}
