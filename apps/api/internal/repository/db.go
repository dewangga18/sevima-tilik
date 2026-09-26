package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	*sql.DB
}

func Connect(ctx context.Context, databaseURL string, seedDemoUsers bool) (*DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Wait for db to be available (retry up to 10 times)
	var pingErr error
	for i := 0; i < 10; i++ {
		pingErr = db.PingContext(ctx)
		if pingErr == nil {
			break
		}
		log.Printf("Waiting for database connection (%d/10)...: %v", i+1, pingErr)
		time.Sleep(2 * time.Second)
	}
	if pingErr != nil {
		return nil, fmt.Errorf("ping db failed: %w", pingErr)
	}

	wrapped := &DB{db}
	if err := wrapped.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	if seedDemoUsers {
		if err := wrapped.SeedDemoUsers(ctx); err != nil {
			return nil, fmt.Errorf("seed demo users: %w", err)
		}
	}
	if err := wrapped.SeedClassrooms(ctx); err != nil {
		return nil, fmt.Errorf("seed classrooms: %w", err)
	}
	if err := wrapped.Seed(ctx); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	if err := wrapped.SeedLearning(ctx); err != nil {
		return nil, fmt.Errorf("seed learning: %w", err)
	}
	return wrapped, nil
}

func (db *DB) Migrate(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		role TEXT NOT NULL,
		grade_level INT DEFAULT 4,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS skills (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		domain TEXT NOT NULL,
		grade_level INT NOT NULL,
		description TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS skill_prerequisites (
		skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
		prereq_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
		PRIMARY KEY (skill_id, prereq_id)
	);

	CREATE TABLE IF NOT EXISTS questions (
		id TEXT PRIMARY KEY,
		skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
		difficulty INT NOT NULL,
		prompt TEXT NOT NULL,
		options JSONB NOT NULL,
		answer_key TEXT NOT NULL,
		explanation TEXT NOT NULL,
		misconception TEXT DEFAULT ''
	);
	CREATE TABLE IF NOT EXISTS question_candidates (
		id TEXT PRIMARY KEY,
		content_hash TEXT NOT NULL,
		source_file TEXT NOT NULL,
		payload JSONB NOT NULL,
		status TEXT NOT NULL CHECK(status IN ('draft','approved','rejected')),
		review_note TEXT NOT NULL,
		registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS assessments (
		id TEXT PRIMARY KEY,
		student_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		grade_level INT NOT NULL,
		status TEXT NOT NULL,
		started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		completed_at TIMESTAMP WITH TIME ZONE
	);

	CREATE TABLE IF NOT EXISTS assessment_items (
		id TEXT PRIMARY KEY,
		assessment_id TEXT NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
		question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
		order_index INT NOT NULL,
		student_answer TEXT DEFAULT '',
		is_correct BOOLEAN,
		answered_at TIMESTAMP WITH TIME ZONE
	);

	CREATE TABLE IF NOT EXISTS skill_evidence (
		id TEXT PRIMARY KEY,
		student_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
		assessment_id TEXT NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
		status TEXT NOT NULL,
		total_answered INT NOT NULL,
		total_correct INT NOT NULL,
		evidence_count INT NOT NULL,
		confidence TEXT NOT NULL,
		is_root_gap BOOLEAN NOT NULL DEFAULT FALSE,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
 ALTER TABLE assessments ADD COLUMN IF NOT EXISTS rule_version TEXT NOT NULL DEFAULT 'legacy-v1';
 ALTER TABLE assessments ADD COLUMN IF NOT EXISTS target_skill_id TEXT NOT NULL DEFAULT '';
 ALTER TABLE assessments ADD COLUMN IF NOT EXISTS revision INT NOT NULL DEFAULT 0;
 ALTER TABLE assessments ADD COLUMN IF NOT EXISTS stop_reason TEXT NOT NULL DEFAULT '';
 ALTER TABLE assessments ADD COLUMN IF NOT EXISTS learning_path JSONB NOT NULL DEFAULT '[]';
 ALTER TABLE assessment_items ADD COLUMN IF NOT EXISTS probe_for_skill_id TEXT NOT NULL DEFAULT '';
 ALTER TABLE skill_evidence ADD COLUMN IF NOT EXISTS related_target_skill_id TEXT NOT NULL DEFAULT '';

 ALTER TABLE questions ADD COLUMN IF NOT EXISTS purpose TEXT NOT NULL DEFAULT 'diagnostic';
 CREATE TABLE IF NOT EXISTS lessons (
  id TEXT PRIMARY KEY, skill_id TEXT NOT NULL UNIQUE REFERENCES skills(id),title TEXT NOT NULL,estimated_minutes INT NOT NULL,steps JSONB NOT NULL,hint TEXT NOT NULL
 );
 CREATE TABLE IF NOT EXISTS learning_sessions (
  id TEXT PRIMARY KEY,student_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,source_assessment_id TEXT NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,skill_id TEXT NOT NULL REFERENCES skills(id),request_id TEXT NOT NULL,rule_version TEXT NOT NULL,stage TEXT NOT NULL,revision INT NOT NULL DEFAULT 0,started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),lesson_completed_at TIMESTAMPTZ,completed_at TIMESTAMPTZ,before_score INT,score INT,outcome TEXT NOT NULL DEFAULT '',stop_reason TEXT NOT NULL DEFAULT '',review_skill_id TEXT NOT NULL DEFAULT '',UNIQUE(student_id,request_id)
 );
 CREATE TABLE IF NOT EXISTS learning_start_requests (
  student_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,request_id TEXT NOT NULL,source_assessment_id TEXT NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,skill_id TEXT NOT NULL REFERENCES skills(id),session_id TEXT NOT NULL REFERENCES learning_sessions(id) ON DELETE CASCADE,PRIMARY KEY(student_id,request_id)
 );
 CREATE TABLE IF NOT EXISTS learning_items (
  id TEXT PRIMARY KEY,session_id TEXT NOT NULL REFERENCES learning_sessions(id) ON DELETE CASCADE,question_id TEXT NOT NULL REFERENCES questions(id),stage TEXT NOT NULL,order_index INT NOT NULL,student_answer TEXT,is_correct BOOLEAN,answered_at TIMESTAMPTZ,UNIQUE(session_id,question_id),UNIQUE(session_id,order_index)
 );
  CREATE TABLE IF NOT EXISTS skill_progress (
   student_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,skill_id TEXT NOT NULL REFERENCES skills(id),source_session_id TEXT NOT NULL REFERENCES learning_sessions(id) ON DELETE CASCADE,score INT NOT NULL,status TEXT NOT NULL,evidence_count INT NOT NULL,correct_count INT NOT NULL,updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),PRIMARY KEY(student_id,skill_id)
  );
  CREATE TABLE IF NOT EXISTS classrooms (
   id TEXT PRIMARY KEY,name TEXT NOT NULL,grade_level INT NOT NULL
  );
  CREATE TABLE IF NOT EXISTS enrollments (
   classroom_id TEXT NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,student_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,PRIMARY KEY(classroom_id,student_id)
  );
  CREATE TABLE IF NOT EXISTS teacher_assignments (
   teacher_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,classroom_id TEXT NOT NULL REFERENCES classrooms(id) ON DELETE CASCADE,PRIMARY KEY(teacher_id,classroom_id)
  );
	CREATE INDEX IF NOT EXISTS idx_enrollments_student ON enrollments(student_id);
	CREATE INDEX IF NOT EXISTS idx_teacher_assignments_class ON teacher_assignments(classroom_id);

	`
	_, err := db.ExecContext(ctx, schema)
	return err
}

func (db *DB) SeedDemoUsers(ctx context.Context) error {
	// Demo accounts use the development-only login endpoint, never a shared password.
	password := make([]byte, 32)
	if _, err := rand.Read(password); err != nil {
		return err
	}
	pwHash, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(password)), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []struct {
		id, email, name, role string
		grade                 int
	}{
		{"u-student-1", "budi@tilik.id", "Budi Santoso", "student", 4},
		{"u-student-new", "budi.baru@tilik.id", "Budi Santoso", "student", 4},
		{"u-student-2", "ani@tilik.id", "Ani Wijaya", "student", 4},
		{"u-student-3", "deni@tilik.id", "Deni Pratama", "student", 4},
		{"u-teacher-1", "siti@tilik.id", "Ibu Siti Rahayu", "teacher", 4},
		{"u-teacher-2", "rahmat@tilik.id", "Pak Rahmat", "teacher", 4},
		{"u-admin-1", "admin@tilik.id", "Admin Tilik", "admin", 0},
	}

	for _, u := range users {
		query := `
			INSERT INTO users (id, email, password_hash, name, role, grade_level)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (email) DO NOTHING
		`
		if _, err := db.ExecContext(ctx, query, u.id, u.email, string(pwHash), u.name, u.role, u.grade); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) SeedClassrooms(ctx context.Context) error {
	classrooms := []struct {
		id, name string
		grade    int
	}{
		{"cls-4a", "Kelas 4A", 4},
		{"cls-4b", "Kelas 4B", 4},
	}
	for _, c := range classrooms {
		if _, err := db.ExecContext(ctx, `INSERT INTO classrooms (id, name, grade_level) VALUES ($1,$2,$3) ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, grade_level=EXCLUDED.grade_level`, c.id, c.name, c.grade); err != nil {
			return err
		}
	}

	enrollments := []struct{ class, student string }{
		{"cls-4a", "u-student-1"},
		{"cls-4a", "u-student-new"},
		{"cls-4a", "u-student-2"},
		{"cls-4a", "u-student-3"},
	}
	for _, e := range enrollments {
		if _, err := db.ExecContext(ctx, `INSERT INTO enrollments (classroom_id, student_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, e.class, e.student); err != nil {
			return err
		}
	}

	// Siti mengajar 4A dan 4B; Rahmat hanya 4B untuk memverifikasi batas akses antar kelas.
	assignments := []struct{ teacher, class string }{
		{"u-teacher-1", "cls-4a"},
		{"u-teacher-1", "cls-4b"},
		{"u-teacher-2", "cls-4b"},
	}
	for _, a := range assignments {
		if _, err := db.ExecContext(ctx, `INSERT INTO teacher_assignments (teacher_id, classroom_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, a.teacher, a.class); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) Seed(ctx context.Context) error {
	// Seed Grade 4 slice skills
	skills := []struct {
		id, name, domain, desc string
		grade                  int
	}{
		{"mul_basic", "Perkalian Dasar", "Bilangan", "Perkalian bilangan cacah dasar 1-10", 3},
		{"div_basic", "Pembagian Dasar", "Bilangan", "Pembagian bilangan cacah dasar dan konsep bagi adil", 3},
		{"frac_rep", "Representasi Pecahan", "Pecahan", "Memahami konsep pecahan bagian dari keutuhan dan gambar", 4},
		{"frac_equiv", "Pecahan Senilai", "Pecahan", "Menentukan dan menyederhanakan pecahan yang senilai", 4},
		{"frac_cmp_same_den", "Perbandingan Pecahan Berpenyebut Sama", "Pecahan", "Membandingkan pecahan dengan pembagi yang sama", 4},
		{"frac_cmp_diff_den", "Perbandingan Pecahan Berbeda Penyebut", "Pecahan", "Membandingkan pecahan dengan penyebut tidak sama", 4},
	}

	for _, s := range skills {
		query := `
			INSERT INTO skills (id, name, domain, grade_level, description)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, description=EXCLUDED.description
		`
		if _, err := db.ExecContext(ctx, query, s.id, s.name, s.domain, s.grade, s.desc); err != nil {
			return err
		}
	}

	// Seed skill prerequisites DAG
	prereqs := []struct {
		skillID, prereqID string
	}{
		{"frac_equiv", "mul_basic"},
		{"frac_equiv", "div_basic"},
		{"frac_equiv", "frac_rep"},
		{"frac_cmp_same_den", "frac_rep"},
		{"frac_cmp_diff_den", "frac_equiv"},
		{"frac_cmp_diff_den", "frac_cmp_same_den"},
	}

	for _, p := range prereqs {
		query := `
			INSERT INTO skill_prerequisites (skill_id, prereq_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
		if _, err := db.ExecContext(ctx, query, p.skillID, p.prereqID); err != nil {
			return err
		}
	}

	// Seed reviewed questions bank for the Grade 4 slice
	type qSeed struct {
		id            string
		skillID       string
		difficulty    int
		prompt        string
		options       []string
		answerKey     string
		explanation   string
		misconception string
	}

	seedQuestions := []qSeed{
		// 1. mul_basic
		{
			id:            "q-mul-1",
			skillID:       "mul_basic",
			difficulty:    1,
			prompt:        "Berapakah hasil dari 4 × 3?",
			options:       []string{"7", "12", "14", "16"},
			answerKey:     "12",
			explanation:   "4 × 3 artinya penjumlahan berulang 4 sebanyak 3 kali: 4 + 4 + 4 = 12.",
			misconception: "Menjumlahkan 4 + 3 = 7 alih-alih mengalikan.",
		},
		{
			id:            "q-mul-2",
			skillID:       "mul_basic",
			difficulty:    1,
			prompt:        "Ibu memiliki 5 kantong permen. Setiap kantong berisi 6 permen. Berapa total permen Ibu?",
			options:       []string{"11", "25", "30", "35"},
			answerKey:     "30",
			explanation:   "Total permen = 5 × 6 = 30 permen.",
			misconception: "Menjumlahkan 5 + 6 = 11.",
		},
		{
			id:            "q-mul-3",
			skillID:       "mul_basic",
			difficulty:    2,
			prompt:        "Berapakah hasil dari 7 × 8?",
			options:       []string{"54", "56", "58", "64"},
			answerKey:     "56",
			explanation:   "7 × 8 = 56.",
			misconception: "Salah menghafal tabel perkalian (misal tertukar dengan 6 × 9 = 54).",
		},

		// 2. div_basic
		{
			id:            "q-div-1",
			skillID:       "div_basic",
			difficulty:    1,
			prompt:        "Berapakah hasil dari 18 ÷ 3?",
			options:       []string{"5", "6", "7", "9"},
			answerKey:     "6",
			explanation:   "18 dibagi rata ke 3 kelompok menghasilkan 6 pada setiap kelompok (karena 6 × 3 = 18).",
			misconception: "Salah membagi (misal 18 ÷ 2 = 9).",
		},
		{
			id:            "q-div-2",
			skillID:       "div_basic",
			difficulty:    1,
			prompt:        "Ada 24 buku yang dibagikan sama rata kepada 4 anak. Berapa buku yang diterima setiap anak?",
			options:       []string{"4", "6", "8", "12"},
			answerKey:     "6",
			explanation:   "24 ÷ 4 = 6 buku per anak.",
			misconception: "Menebak 8 karena 3 × 8 = 24 tanpa melihat pembagi 4.",
		},
		{
			id:            "q-div-3",
			skillID:       "div_basic",
			difficulty:    2,
			prompt:        "Berapakah hasil dari 42 ÷ 6?",
			options:       []string{"6", "7", "8", "9"},
			answerKey:     "7",
			explanation:   "42 ÷ 6 = 7 (karena 7 × 6 = 42).",
			misconception: "Menebak 6 karena 6 × 6 = 36.",
		},

		// 3. frac_rep
		{
			id:            "q-rep-1",
			skillID:       "frac_rep",
			difficulty:    1,
			prompt:        "Sebuah pizza dipotong menjadi 4 bagian sama besar. Budi memakan 1 potong. Berapa bagian pizza yang dimakan Budi?",
			options:       []string{"1/4", "1/3", "4/1", "3/4"},
			answerKey:     "1/4",
			explanation:   "Pecahan menunjukkan bagian yang diambil (pembilang = 1) dari seluruh bagian (penyebut = 4), yaitu 1/4.",
			misconception: "Menulis sisa potongan sebagai penyebut (1/3) atau membalik pembilang/penyebut (4/1).",
		},
		{
			id:            "q-rep-2",
			skillID:       "frac_rep",
			difficulty:    1,
			prompt:        "Pada pecahan 3/8, angka 8 disebut sebagai...",
			options:       []string{"Pembilang", "Penyebut", "Sisa", "Hasil bagi"},
			answerKey:     "Penyebut",
			explanation:   "Pada pecahan a/b, angka atas (3) adalah pembilang dan angka bawah (8) adalah penyebut.",
			misconception: "Tertukar antara istilah pembilang dan penyebut.",
		},
		{
			id:            "q-rep-3",
			skillID:       "frac_rep",
			difficulty:    2,
			prompt:        "Dari 10 buah apel di keranjang, 4 buah berwarna merah. Berapakah pecahan apel merah dari seluruh apel?",
			options:       []string{"4/10", "4/6", "6/10", "10/4"},
			answerKey:     "4/10",
			explanation:   "Jumlah apel merah adalah 4 dari total 10 apel, sehingga pecahannya adalah 4/10.",
			misconception: "Membandingkan apel merah dengan apel hijau saja (4/6).",
		},

		// 4. frac_equiv
		{
			id:            "q-eq-1",
			skillID:       "frac_equiv",
			difficulty:    1,
			prompt:        "Manakah pecahan berikut yang senilai dengan 1/2?",
			options:       []string{"2/4", "2/3", "1/4", "3/5"},
			answerKey:     "2/4",
			explanation:   "Jika pembilang dan penyebut 1/2 dikalikan 2, maka: (1×2)/(2×2) = 2/4.",
			misconception: "Mengira pecahan dengan angka berbeda tidak bisa bernilai sama.",
		},
		{
			id:            "q-eq-2",
			skillID:       "frac_equiv",
			difficulty:    2,
			prompt:        "Pecahan yang senilai dengan 2/3 adalah...",
			options:       []string{"4/6", "3/4", "4/9", "5/6"},
			answerKey:     "4/6",
			explanation:   "Mengalikan pembilang dan penyebut dengan 2: (2×2)/(3×2) = 4/6.",
			misconception: "Menambahkan angka yang sama ke atas dan bawah: (2+2)/(3+2) = 4/5 atau salah kali.",
		},
		{
			id:            "q-eq-3",
			skillID:       "frac_equiv",
			difficulty:    2,
			prompt:        "Bentuk paling sederhana dari pecahan 6/8 adalah...",
			options:       []string{"3/4", "2/3", "3/8", "1/2"},
			answerKey:     "3/4",
			explanation:   "Bagi pembilang dan penyebut dengan FPB-nya (2): 6÷2 = 3, 8÷2 = 4, jadi 3/4.",
			misconception: "Salah membagi atau hanya membagi salah satu komponen.",
		},

		// 5. frac_cmp_same_den
		{
			id:            "q-cmp-same-1",
			skillID:       "frac_cmp_same_den",
			difficulty:    1,
			prompt:        "Tanda yang tepat untuk membandingkan 2/5 ... 4/5 adalah...",
			options:       []string{"<", ">", "=", "≤"},
			answerKey:     "<",
			explanation:   "Jika penyebutnya sama (5), pecahan dengan pembilang lebih kecil bernilai lebih kecil. 2 < 4, maka 2/5 < 4/5.",
			misconception: "Mengira angka lebih besar di pembilang berarti potongan lebih sedikit.",
		},
		{
			id:            "q-cmp-same-2",
			skillID:       "frac_cmp_same_den",
			difficulty:    1,
			prompt:        "Manakah pecahan yang paling besar?",
			options:       []string{"1/7", "3/7", "5/7", "2/7"},
			answerKey:     "5/7",
			explanation:   "Pada penyebut yang sama (7), pembilang terbesar adalah 5, jadi 5/7 adalah yang paling besar.",
			misconception: "Memilih pembilang terkecil karena salah konsep pembagian.",
		},
		{
			id:            "q-cmp-same-3",
			skillID:       "frac_cmp_same_den",
			difficulty:    2,
			prompt:        "Urutan pecahan dari yang terkecil: 6/11, 2/11, 8/11 adalah...",
			options:       []string{"2/11, 6/11, 8/11", "8/11, 6/11, 2/11", "2/11, 8/11, 6/11", "6/11, 2/11, 8/11"},
			answerKey:     "2/11, 6/11, 8/11",
			explanation:   "Karena penyebut sama (11), urutkan pembilang dari yang terkecil: 2 < 6 < 8.",
			misconception: "Mengurutkan terbalik (dari terbesar ke terkecil).",
		},

		// 6. frac_cmp_diff_den
		{
			id:            "q-cmp-diff-1",
			skillID:       "frac_cmp_diff_den",
			difficulty:    2,
			prompt:        "Tanda yang tepat untuk membandingkan 1/2 ... 1/4 adalah...",
			options:       []string{">", "<", "=", "≥"},
			answerKey:     ">",
			explanation:   "Satu benda dibagi 2 potongannya lebih besar daripada dibagi 4 bagian. Atau samakan penyebut: 1/2 = 2/4 > 1/4.",
			misconception: "Melihat angka 4 > 2 lalu menyimpulkan 1/4 lebih besar (common whole number bias).",
		},
		{
			id:            "q-cmp-diff-2",
			skillID:       "frac_cmp_diff_den",
			difficulty:    2,
			prompt:        "Tanda yang tepat untuk membandingkan 2/3 ... 3/4 adalah...",
			options:       []string{"<", ">", "=", "≤"},
			answerKey:     "<",
			explanation:   "Samakan penyebut ke KPK 12: 2/3 = 8/12 dan 3/4 = 9/12. Karena 8/12 < 9/12, maka 2/3 < 3/4.",
			misconception: "Melihat selisih pembilang dan penyebut sama-sama 1 lalu mengira nilainya sama (=).",
		},
		{
			id:            "q-cmp-diff-3",
			skillID:       "frac_cmp_diff_den",
			difficulty:    2,
			prompt:        "Manakah pernyataan yang benar?",
			options:       []string{"3/5 > 1/2", "1/3 > 1/2", "2/5 = 1/2", "3/4 < 2/4"},
			answerKey:     "3/5 > 1/2",
			explanation:   "3/5 = 6/10 dan 1/2 = 5/10. Karena 6/10 > 5/10, maka 3/5 > 1/2.",
			misconception: "Whole number bias pada penyebut.",
		},
	}

	for _, q := range seedQuestions {
		optsJSON, err := json.Marshal(q.options)
		if err != nil {
			return err
		}

		query := `
			INSERT INTO questions (id, skill_id, difficulty, prompt, options, answer_key, explanation, misconception)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET
				prompt=EXCLUDED.prompt,
				options=EXCLUDED.options,
				answer_key=EXCLUDED.answer_key,
				explanation=EXCLUDED.explanation,
				misconception=EXCLUDED.misconception
		`
		if _, err := db.ExecContext(ctx, query, q.id, q.skillID, q.difficulty, q.prompt, optsJSON, q.answerKey, q.explanation, q.misconception); err != nil {
			return err
		}
	}

	return nil
}
