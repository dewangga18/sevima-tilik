package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sevima/tilik-api/internal/domain"
)

func learningLessons() []domain.Lesson {
	return []domain.Lesson{
		{ID: "lesson-mul", SkillID: "mul_basic", Title: "Perkalian sebagai kelompok yang sama", EstimatedMinutes: 2, Hint: "Hitung banyak kelompok dan isi tiap kelompok, lalu kalikan.", Steps: []domain.LessonStep{
			{Heading: "Lihat kelompoknya", Body: "Perkalian membantu menghitung beberapa kelompok dengan isi sama.", Example: "3 kelompok, masing-masing 4 benda."},
			{Heading: "Hubungkan dengan penjumlahan", Body: "Jumlahkan isi tiap kelompok untuk melihat makna perkalian.", Example: "4 + 4 + 4 = 12; jadi 3 × 4 = 12."},
			{Heading: "Periksa hasilnya", Body: "Banyak kelompok dikali isi kelompok menghasilkan jumlah seluruh benda.", Example: "5 kotak berisi 2 pensil: 5 × 2 = 10 pensil."}}},
		{ID: "lesson-div", SkillID: "div_basic", Title: "Membagi sama banyak", EstimatedMinutes: 2, Hint: "Cari perkalian yang menghasilkan jumlah seluruh benda.", Steps: []domain.LessonStep{
			{Heading: "Bagikan dengan adil", Body: "Pembagian sama banyak memberi tiap kelompok jumlah yang sama.", Example: "12 stiker dibagikan kepada 3 anak."},
			{Heading: "Cari isi tiap bagian", Body: "Tanyakan berapa benda yang diterima setiap anak.", Example: "12 ÷ 3 = 4; setiap anak mendapat 4 stiker."},
			{Heading: "Periksa dengan perkalian", Body: "Kalikan jumlah penerima dengan isi tiap bagian.", Example: "3 × 4 = 12; hasil pembagian sesuai jumlah awal."}}},
		{ID: "lesson-rep", SkillID: "frac_rep", Title: "Pecahan adalah bagian dari satu utuh", EstimatedMinutes: 2, Hint: "Penyebut menghitung semua bagian sama besar; pembilang menghitung bagian yang dipilih.", Steps: []domain.LessonStep{
			{Heading: "Bagi satu utuh", Body: "Pecahan menggunakan bagian yang sama besar.", Example: "Satu roti dibagi menjadi 4 bagian sama besar."},
			{Heading: "Hitung bagian yang dipilih", Body: "Pembilang menunjukkan bagian yang dipilih; penyebut menunjukkan seluruh bagian.", Example: "Jika 1 bagian diambil, pecahannya 1/4."},
			{Heading: "Baca pecahan", Body: "Perhatikan jumlah bagian, bukan ukuran gambar saja.", Example: "3 dari 5 bagian sama besar berarti 3/5."}}},
		{ID: "lesson-eq", SkillID: "frac_equiv", Title: "Pecahan senilai, bentuk berbeda", EstimatedMinutes: 2, Hint: "Kalikan pembilang dan penyebut dengan bilangan yang sama.", Steps: []domain.LessonStep{
			{Heading: "Nilainya tetap", Body: "Pecahan senilai menggambarkan bagian yang sama dari satu utuh.", Example: "1/2 dan 2/4 sama-sama setengah."},
			{Heading: "Ubah kedua bilangan", Body: "Kalikan pembilang dan penyebut dengan bilangan positif yang sama.", Example: "1/2 = (1 × 3)/(2 × 3) = 3/6."},
			{Heading: "Periksa pasangan", Body: "Mengubah hanya pembilang atau penyebut mengubah nilai pecahan.", Example: "2/3 = 4/6; keduanya dikali 2."}}},
		{ID: "lesson-same", SkillID: "frac_cmp_same_den", Title: "Membandingkan bagian yang sama besar", EstimatedMinutes: 2, Hint: "Jika penyebut sama, bandingkan pembilangnya.", Steps: []domain.LessonStep{
			{Heading: "Periksa penyebut", Body: "Penyebut yang sama berarti setiap bagian sama besar pada satu utuh yang sama.", Example: "2/7 dan 5/7 sama-sama memakai bagian per tujuh."},
			{Heading: "Bandingkan pembilang", Body: "Lebih banyak bagian berarti pecahan lebih besar.", Example: "2/7 < 5/7 karena 2 < 5."},
			{Heading: "Pilih tanda", Body: "Gunakan < untuk lebih kecil, > untuk lebih besar, dan = jika sama.", Example: "4/9 > 1/9; 3/8 = 3/8."}}},
		{ID: "lesson-diff", SkillID: "frac_cmp_diff_den", Title: "Samakan bagian sebelum membandingkan", EstimatedMinutes: 2, Hint: "Ubah menjadi pecahan senilai dengan penyebut sama, lalu bandingkan pembilang.", Steps: []domain.LessonStep{
			{Heading: "Bagian belum sama", Body: "Penyebut berbeda berarti ukuran bagian berbeda, jadi pembilang saja belum cukup.", Example: "1/2 dan 1/3 memiliki pembilang sama, tetapi nilainya berbeda."},
			{Heading: "Cari penyebut bersama", Body: "Ubah kedua pecahan menjadi pecahan senilai dengan penyebut sama.", Example: "1/2 = 3/6; 1/3 = 2/6."},
			{Heading: "Bandingkan hasil", Body: "Setelah penyebut sama, bandingkan pembilangnya.", Example: "3/6 > 2/6, jadi 1/2 > 1/3."}}},
	}
}

func learningQuestion(skill, purpose string, index int) domain.Question {
	a, b := index+2, index+3
	if purpose == "reassessment" {
		a, b = 8-index, 2+index%2
	}
	q := domain.Question{ID: fmt.Sprintf("learn-%s-%s-%d", purpose, skill, index+1), SkillID: skill, Difficulty: 1}
	if index >= 3 {
		q.Difficulty = 2
	}
	switch skill {
	case "mul_basic":
		correct := a * b
		q.Prompt = fmt.Sprintf("Ada %d kotak. Tiap kotak berisi %d kelereng. Berapa jumlah seluruh kelereng?", a, b)
		q.AnswerKey = strconv.Itoa(correct)
		q.Options = []string{q.AnswerKey, strconv.Itoa(correct + 1), strconv.Itoa(correct - b), strconv.Itoa(correct + b)}
		q.Explanation = fmt.Sprintf("Jumlah kelompok × isi tiap kelompok: %d × %d = %d.", a, b, correct)
	case "div_basic":
		q.Prompt = fmt.Sprintf("%d stiker dibagikan sama banyak kepada %d anak. Berapa stiker yang diterima tiap anak?", a*b, b)
		q.AnswerKey = strconv.Itoa(a)
		q.Options = []string{q.AnswerKey, strconv.Itoa(a + 1), strconv.Itoa(a - 1), strconv.Itoa(a + 2)}
		q.Explanation = fmt.Sprintf("%d ÷ %d = %d. Periksa: %d × %d = %d.", a*b, b, a, b, a, a*b)
	case "frac_rep":
		part, total := index+1, index+5
		if purpose == "reassessment" {
			part, total = index+2, index+9
		}
		q.Prompt = fmt.Sprintf("Satu pita dibagi menjadi %d bagian sama panjang. %d bagian diberi warna. Pecahan bagian berwarna adalah...", total, part)
		q.AnswerKey = fmt.Sprintf("%d/%d", part, total)
		q.Options = []string{q.AnswerKey, fmt.Sprintf("%d/%d", total, part), fmt.Sprintf("%d/%d", part, total+1), fmt.Sprintf("%d/%d", part+1, total)}
		q.Explanation = fmt.Sprintf("Pembilang = %d bagian berwarna; penyebut = %d seluruh bagian. Jadi %s.", part, total, q.AnswerKey)
	case "frac_equiv":
		numerator, denominator, factor := index+1, index+3, 2+index%3
		if purpose == "reassessment" {
			numerator, denominator, factor = index+2, index+5, 3+index%3
		}
		q.Prompt = fmt.Sprintf("Pembilang dan penyebut %d/%d sama-sama dikali %d. Pecahan senilainya adalah...", numerator, denominator, factor)
		q.AnswerKey = fmt.Sprintf("%d/%d", numerator*factor, denominator*factor)
		q.Options = []string{q.AnswerKey, fmt.Sprintf("%d/%d", numerator, denominator*factor), fmt.Sprintf("%d/%d", numerator*factor, denominator), fmt.Sprintf("%d/%d", numerator*factor+1, denominator*factor)}
		q.Explanation = fmt.Sprintf("Kalikan keduanya: (%d × %d)/(%d × %d) = %s.", numerator, factor, denominator, factor, q.AnswerKey)
	case "frac_cmp_same_den", "frac_cmp_diff_den":
		n1, d1, n2, d2 := a, a+4, a+2, a+4
		if purpose == "reassessment" {
			n1, d1, n2, d2 = a, a+6, a+1, a+6
		}
		if skill == "frac_cmp_diff_den" {
			n1, d1, n2, d2 = 1, a, 1, b
			if index >= 3 {
				n1, d1, n2, d2 = a, a+3, a-1, a+4
				if purpose == "reassessment" {
					d1, d2 = a+5, a+6
				}
			}
		}
		if index%2 == 1 {
			n1, n2 = n2, n1
			d1, d2 = d2, d1
		}
		q.Prompt = fmt.Sprintf("Pilih tanda <, >, atau = untuk %d/%d ... %d/%d.", n1, d1, n2, d2)
		q.AnswerKey = "="
		if n1*d2 < n2*d1 {
			q.AnswerKey = "<"
		}
		if n1*d2 > n2*d1 {
			q.AnswerKey = ">"
		}
		q.Options = []string{">", "<", "=", "Tidak dapat dibandingkan"}
		q.Explanation = fmt.Sprintf("Dengan penyebut bersama %d: %d/%d = %d/%d dan %d/%d = %d/%d. Jadi tanda yang tepat %s.", d1*d2, n1, d1, n1*d2, d1*d2, n2, d2, n2*d1, d1*d2, q.AnswerKey)
	}
	// Vary answer positions; do not teach a fixed correct-option pattern.
	rotation := index % 4
	q.Options = append(q.Options[rotation:], q.Options[:rotation]...)
	return q
}

func (db *DB) SeedLearning(ctx context.Context) error {
	for _, lesson := range learningLessons() {
		steps, err := json.Marshal(lesson.Steps)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, `INSERT INTO lessons(id,skill_id,title,estimated_minutes,steps,hint) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(id) DO NOTHING`, lesson.ID, lesson.SkillID, lesson.Title, lesson.EstimatedMinutes, steps, lesson.Hint)
		if err != nil {
			return err
		}
		for _, purpose := range []string{"practice", "reassessment"} {
			for i := 0; i < 6; i++ {
				q := learningQuestion(lesson.SkillID, purpose, i)
				options, err := json.Marshal(q.Options)
				if err != nil {
					return err
				}
				_, err = db.ExecContext(ctx, `INSERT INTO questions(id,skill_id,difficulty,prompt,options,answer_key,explanation,purpose) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(id) DO NOTHING`, q.ID, q.SkillID, q.Difficulty, q.Prompt, options, q.AnswerKey, q.Explanation, purpose)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
