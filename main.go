package main

import (
	"embed"
	"encoding/gob"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"  // ← добавить

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type CouponData struct {
	Date     string
	Time     string
	Place    string
	UserName string
	CreatedAt string
}

type MusicQuestion struct {
	ID       int
	Title    string
	Artist   string
	AudioURL string
	Options  []string
	Correct  int
}

var musicQuestions = []MusicQuestion{
	{
		ID:       0,
		Title:    "🎵 Угадай песню #1",
		Artist:   "Подсказка: это точно не газан 67",
		AudioURL: "/static/music/track1.mp3",
		Options:  []string{"CUPSIZE - по барабану ", "YARUSHIN - ZOMBIE", "GAZAN - 67", "CUPSIZE - пока пока", "ssshhhiiittt - танцы", "CUPSIZE - ЗПП"},
		Correct:  0,
	},
	{
		ID:       1,
		Title:    "🎵 Угадай песню #2",
		Artist:   "Подсказка: я ее не придумал",
		AudioURL: "/static/music/track2.mp3",
		Options:  []string{"Серега пират - прости я не знаю", "KSB music - дота ", "KSB music - сахарная пудра", "KSB music - я вытащу тебя со дна", "серега пират - ну где моя нога", "Tokyo Ghoul - unravel"},
		Correct:  2,
	},
	{
		ID:       2,
		Title:    "🎵 Угадай песню #3",
		Artist:   "Подсказка: точна зарубежная",
		AudioURL: "/static/music/track3.mp3",
		Options:  []string{"girl in red - we fell in love in october", "TV girls - Song About Me", "Mindless Self Indulgence - Mastermind", "Alex G - Sarah", "And one - Military fashion show", " Mikky - Just The Way You Are"},
		Correct:  1,
	},
	{
		ID:       3,
		Title:    "🎵 Угадай песню #4",
		Artist:   "Подсказка:Я русский",
		AudioURL: "/static/music/track4.mp3",
		Options:  []string{"DEAD BLONDE - СНЕГ РАСТАЯЛ НА ПЛЕЧАХ", "МС Вспышкин и Никифоровна - Колбасный цех 3 (шишки)", "Песня о любви — Уральские Пельмени", "Нейромонах Феофан - Притоптать", "GSPD - Никому не говори", "Егор Крид - будильник"},
		Correct:  4,
	},
	{
		ID:       4,
		Title:    "🎵 Угадай песню #5",
		Artist:   "Подсказка: я схожу с ума пока придумываю какие песенки вставить",
		AudioURL: "/static/music/track5.mp3",
		Options:  []string{"ComedoZ - Ямайка", "Женский гимн", "Ляпис Трубецкой — Капитал", "Monetochka - Мартовские коты", "Monetochka - кис кис", "Филипп Киркоров - Давно все хорошо "},
		Correct:  3,
	},
	{
		ID:       5,
		Title:    "🎵 Угадай песню #6",
		Artist:   "Подсказка: пук пук пук бле бле бле ",
		AudioURL: "/static/music/track6.mp3",
		Options:  []string{"Noize Mc - Заебались!", "GSPD - Робот", "Valentin Strykalo - бу бу ", "плм - я фудкортница", "конец солнечных дней - шалайла бум бум", "плм - кхм кхм"},
		Correct:  5,
	},
}

type User struct {
	Name     string
	Birthday string
	Wish     string
}

type Photo struct {
	ID          int
	ImageURL    string
	Caption     string
	Description string
}

type Question struct {
	ID      int
	Text    string
	Options []string
	Correct int
	Comment string
}

type MusicSession struct {
	CurrentQuestion int
	CorrectCount    int // просто счетчик правильных ответов
	TotalAnswered   int // сколько вопросов已回答
}

const targetBirthday = "2026-08-09"

//go:embed templates/* static/*
var webFiles embed.FS

var store = sessions.NewCookieStore([]byte("super-secret-key-2026"))

var photos = []Photo{
	{
		ID:          0,
		ImageURL:    "/static/images/placeholder1.jpg",
		Caption:     "✨ Самая красивая ✨",
		Description: "Твоя улыбка освещает этот мир",
	},
	{
		ID:          1,
		ImageURL:    "/static/images/placeholder2.jpg",
		Caption:     "🌟 Самая сильная 🌟",
		Description: " Можешь решить все своми руками ",
	},
	{
		ID:          2,
		ImageURL:    "/static/images/placeholder3.jpg",
		Caption:     "💫 Самая добрая 💫",
		Description: "Твоё сердце согревает всех вокруг",
	},
}

var questions = []Question{
	{
		ID:      0,
		Text:    "Продолжи фразу - Тлен, ужас, страх, уныние, отчаяние, ...",
		Options: []string{"кладбище, скелеты", "голод, смерть", "радость, веселье", "Марина, Юра"},
		Correct: 1,
		Comment: "Правильно! Саша обожает голод, смерть",
	},
	{
		ID:      1,
		Text:    "Какое блюдо Саша может есть каждый день?",
		Options: []string{"Пельмеши", "Жареное пюре", "Карбонара", "Обычный чизбургер"},
		Correct: 2,
		Comment: "Паста — это любовь навсегда",
	},
	{
		ID:      2,
		Text:    "Сашино любимое хобби",
		Options: []string{"Кататься автостопом", "Жить в коммуне", "Ничего не делать", "Играть в роблокс"},
		Correct: 2,
		Comment: "Ха-ха, Саша может ничего не делать часами!",
	},
	{
		ID:      3,
		Text:    "Из чего делаются кошки?",
		Options: []string{"из мяса рыбы", "Что за странный вопрос придумала Марина", "из сахарной ваты", "из шерсти"},
		Correct: 0,
		Comment: "мясо",
	},
	{
		ID:      4,
		Text:    "Сашина главная мечта?",
		Options: []string{"Стать пиратом и грабить работяг", "Уйти в монастырь", "Иметь все психические отклонения", "Стать фудкортницей"},
		Correct: 1,
		Comment: "",
	},
	{
		ID:      5,
		Text:    "Кого Саша любит больше всего",
		Options: []string{"одногрупников", "22 гробовозку", "виноград", "кошечек"},
		Correct: 3,
		Comment: "Лето — время тепла и солнца! ☀️",
	},
}

func main() {
	gob.Register(map[int]bool{})
	gob.Register(User{})
	gob.Register(MusicSession{}) // ← добавить эту строку
	gob.Register(map[string]int{}) // ← добавляем регистрацию map[string]int
	gob.Register(CouponData{})  // ← добавить
	r := mux.NewRouter()

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
// В main() добавь:
r.HandleFunc("/coupon", couponHandler).Methods("GET")
r.HandleFunc("/coupon-form", couponFormHandler).Methods("GET")
r.HandleFunc("/coupon-submit", couponSubmitHandler).Methods("POST")
r.HandleFunc("/coupon-receipt", couponReceiptHandler).Methods("GET")
	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(webFiles)))

	r.HandleFunc("/letter", letterHandler).Methods("GET")
	r.HandleFunc("/open-letter", openLetterHandler).Methods("GET")
	// Музыкальный квиз
	r.HandleFunc("/music", musicIndexHandler).Methods("GET")
	r.HandleFunc("/music-question/{id}", musicQuestionHandler).Methods("GET")
	r.HandleFunc("/music-answer", musicAnswerHandler).Methods("POST")
	r.HandleFunc("/music-result", musicResultHandler).Methods("GET")
	r.HandleFunc("/reset-music", resetMusicHandler).Methods("GET")

	// Основной квиз и дашборд
	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/register", registerHandler).Methods("POST")
	r.HandleFunc("/dashboard", dashboardHandler).Methods("GET")
	r.HandleFunc("/reset-quiz", resetQuizHandler).Methods("GET")
	r.HandleFunc("/question/{id}", questionHandler).Methods("GET")
	r.HandleFunc("/answer", answerHandler).Methods("POST")
	r.HandleFunc("/result", resultHandler).Methods("GET")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")

	log.Println("🚀 Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "quiz")
	if session.Values["user"] != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	tmpl := parseTemplate("index.html")
	tmpl.Execute(w, nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.FormValue("name")
	birthday := r.FormValue("birthday")
	wish := r.FormValue("wish")

	if name == "" || birthday == "" || wish == "" {
		tmpl := parseTemplate("index.html")
		tmpl.Execute(w, map[string]interface{}{
			"Error":    "Пожалуйста, заполните все поля!",
			"Name":     name,
			"Birthday": birthday,
			"Wish":     wish,
		})
		return
	}

	if birthday != targetBirthday {
		tmpl := parseTemplate("index.html")
		tmpl.Execute(w, map[string]interface{}{
			"Error":    "⛔ Доступ запрещён! Этот тест только для Саши!",
			"Name":     name,
			"Birthday": birthday,
			"Wish":     wish,
		})
		return
	}

	session, err := store.Get(r, "quiz")
	if err != nil {
		log.Printf("Ошибка получения сессии: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	session.Values["user"] = User{
		Name:     name,
		Birthday: birthday,
		Wish:     wish,
	}
	session.Values["results"] = make(map[int]bool)

	err = session.Save(r, w)
	if err != nil {
		log.Printf("Ошибка сохранения сессии: %v", err)
		http.Error(w, "Ошибка сохранения сессии", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Пользователь %s успешно зарегистрирован", name)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, ok := session.Values["user"].(User)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Результаты основного квиза
	correct, total := getQuizResults(session)
	results, _ := session.Values["results"].(map[int]bool)
	isQuizCompleted := len(results) == len(questions)

	// Результаты музыкального квиза - теперь простые поля
	musicCorrect := 0
	musicTotal := 0
	isMusicCompleted := false
	
	if val, ok := session.Values["music_correct"].(int); ok {
		musicCorrect = val
	}
	if val, ok := session.Values["music_total"].(int); ok {
		musicTotal = val
	}
	isMusicCompleted = musicTotal > 0
	
	// Если нет сохраненных результатов, но есть musicData
	if !isMusicCompleted {
		if musicData, ok := session.Values["music"].(MusicSession); ok {
			if musicData.TotalAnswered > 0 {
				musicCorrect = musicData.CorrectCount
				musicTotal = len(musicQuestions)
				isMusicCompleted = true
				
				session.Values["music_correct"] = musicCorrect
				session.Values["music_total"] = musicTotal
				session.Save(r, w)
			}
		}
	}

	data := struct {
		User               User
		Photos             []Photo
		CorrectAnswers     int
		TotalQuestions     int
		IsQuizCompleted    bool
		MusicCorrect       int
		MusicTotal         int
		IsMusicCompleted   bool
	}{
		User:               user,
		Photos:             photos,
		CorrectAnswers:     correct,
		TotalQuestions:     total,
		IsQuizCompleted:    isQuizCompleted,
		MusicCorrect:       musicCorrect,
		MusicTotal:         musicTotal,
		IsMusicCompleted:   isMusicCompleted,
	}

	tmpl := parseTemplate("dashboard.html")
	tmpl.Execute(w, data)
}

func resetQuizHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session.Values["results"] = make(map[int]bool)
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Ошибка сохранения сессии при сбросе: %v", err)
	}

	http.Redirect(w, r, "/question/0", http.StatusSeeOther)
}

func questionHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id < 0 || id >= len(questions) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Передаём в шаблон вопрос и общее количество
	data := struct {
		Question   Question
		Total      int
		Current    int
	}{
		Question: questions[id],
		Total:    len(questions),
		Current:  id + 1,
	}

	tmpl := parseTemplate("question.html")
	tmpl.Execute(w, data)
}

func answerHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	qID, err := strconv.Atoi(r.FormValue("qid"))
	if err != nil {
		http.Redirect(w, r, "/question/0", http.StatusSeeOther)
		return
	}

	choice, err := strconv.Atoi(r.FormValue("choice"))
	if err != nil {
		http.Redirect(w, r, "/question/"+strconv.Itoa(qID), http.StatusSeeOther)
		return
	}

	if session.Values["results"] == nil {
		session.Values["results"] = make(map[int]bool)
	}
	results := session.Values["results"].(map[int]bool)

	isCorrect := choice == questions[qID].Correct
	results[qID] = isCorrect
	session.Values["results"] = results
	session.Save(r, w)

	nextID := qID + 1
	if nextID >= len(questions) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/question/"+strconv.Itoa(nextID), http.StatusSeeOther)
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "quiz")
	session.Values["user"] = nil
	session.Values["results"] = nil
	session.Values["music"] = nil
	session.Values["music_results"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getQuizResults(session *sessions.Session) (correct, total int) {
	results, ok := session.Values["results"].(map[int]bool)
	if !ok || results == nil {
		return 0, len(questions)
	}

	correct = 0
	for _, isCorrect := range results {
		if isCorrect {
			correct++
		}
	}
	return correct, len(questions)
}

func parseTemplate(name string) *template.Template {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}
	return template.Must(template.New(name).Funcs(funcMap).ParseFS(webFiles, "templates/"+name))
}

// ========== МУЗЫКАЛЬНЫЙ КВИЗ ==========

func musicIndexHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	musicData := MusicSession{
		CurrentQuestion: 0,
		CorrectCount:    0,
		TotalAnswered:   0,
	}
	session.Values["music"] = musicData
	session.Values["music_results"] = nil
	session.Save(r, w)

	tmpl := parseTemplate("music.html")
	tmpl.Execute(w, nil)
}

func musicQuestionHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	musicData, ok := session.Values["music"].(MusicSession)
	if !ok {
		musicData = MusicSession{
			CurrentQuestion: 0,
			CorrectCount:    0,
			TotalAnswered:   0,
		}
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id < 0 || id >= len(musicQuestions) {
		// Если уже ответили на все вопросы - показываем результат
		if musicData.TotalAnswered >= len(musicQuestions) {
			http.Redirect(w, r, "/music-result", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/music-question/"+strconv.Itoa(musicData.CurrentQuestion), http.StatusSeeOther)
		}
		return
	}

	// Если уже отвечали на этот вопрос, переходим дальше
	if id < musicData.TotalAnswered {
		nextID := id + 1
		if nextID >= len(musicQuestions) {
			http.Redirect(w, r, "/music-result", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/music-question/"+strconv.Itoa(nextID), http.StatusSeeOther)
		}
		return
	}

	musicData.CurrentQuestion = id
	session.Values["music"] = musicData
	session.Save(r, w)

	data := struct {
		Question MusicQuestion
		Total    int
		Current  int
	}{
		Question: musicQuestions[id],
		Total:    len(musicQuestions),
		Current:  id + 1,
	}

	tmpl := parseTemplate("music_question.html")
	tmpl.Execute(w, data)
}

func musicAnswerHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	musicData, ok := session.Values["music"].(MusicSession)
	if !ok {
		musicData = MusicSession{
			CurrentQuestion: 0,
			CorrectCount:    0,
			TotalAnswered:   0,
		}
	}

	r.ParseForm()
	qID, err := strconv.Atoi(r.FormValue("qid"))
	if err != nil {
		http.Redirect(w, r, "/music", http.StatusSeeOther)
		return
	}

	choice, err := strconv.Atoi(r.FormValue("choice"))
	if err != nil {
		log.Printf("Ошибка преобразования choice: %v", err)
		http.Redirect(w, r, "/music-question/"+strconv.Itoa(qID), http.StatusSeeOther)
		return
	}

	isCorrect := choice == musicQuestions[qID].Correct
	
	musicData.TotalAnswered++
	if isCorrect {
		musicData.CorrectCount++
	}
	
	session.Values["music"] = musicData
	
	// Сохраняем простые значения вместо map
	session.Values["music_correct"] = musicData.CorrectCount
	session.Values["music_total"] = len(musicQuestions)
	
	log.Printf("🎵 Вопрос %d: выбрано %d, правильно %d → %v (всего правильно: %d, отвечено: %d)", 
		qID, choice, musicQuestions[qID].Correct, isCorrect, musicData.CorrectCount, musicData.TotalAnswered)
	
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Ошибка сохранения сессии: %v", err)
	}

	nextID := qID + 1
	if nextID >= len(musicQuestions) {
		http.Redirect(w, r, "/music-result", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/music-question/"+strconv.Itoa(nextID), http.StatusSeeOther)
	}
}

func musicResultHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Получаем данные из сессии
	musicData, ok := session.Values["music"].(MusicSession)
	if !ok {
		musicData = MusicSession{
			CurrentQuestion: 0,
			CorrectCount:    0,
			TotalAnswered:   0,
		}
	}
	
	// Используем данные из сессии
	correctCount := 0
	totalQuestions := len(musicQuestions)
	
	// Пробуем получить из сессии
	if val, ok := session.Values["music_correct"].(int); ok {
		correctCount = val
	} else {
		correctCount = musicData.CorrectCount
	}
	
	log.Printf("📊 Итог: %d правильных из %d (отвечено на %d вопросов)", 
		correctCount, totalQuestions, musicData.TotalAnswered)

	data := struct {
		Correct int
		Total   int
	}{
		Correct: correctCount,
		Total:   totalQuestions,
	}

	tmpl := parseTemplate("music_result.html")
	tmpl.Execute(w, data)
}

func resetMusicHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	musicData := MusicSession{
		CurrentQuestion: 0,
		CorrectCount:    0,
		TotalAnswered:   0,
	}
	session.Values["music"] = musicData
	session.Values["music_correct"] = 0
	session.Values["music_total"] = 0
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Ошибка сохранения сессии при сбросе: %v", err)
	}
	
	log.Printf("🔄 Музыкальный квиз сброшен")

	http.Redirect(w, r, "/music", http.StatusSeeOther)
}

func letterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📨 letterHandler: начало обработки")
	
	session, err := store.Get(r, "quiz")
	if err != nil {
		log.Printf("❌ letterHandler: ошибка получения сессии: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	log.Println("✅ letterHandler: сессия получена")
	
	if session.Values["user"] == nil {
		log.Println("❌ letterHandler: пользователь не авторизован")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	log.Println("✅ letterHandler: пользователь авторизован")
	
	user, ok := session.Values["user"].(User)
	if !ok {
		log.Printf("❌ letterHandler: ошибка преобразования user: %v", session.Values["user"])
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	log.Printf("✅ letterHandler: пользователь %s", user.Name)
	
	data := struct {
		User User
	}{
		User: user,
	}
	
	log.Println("📨 letterHandler: парсим шаблон letter.html")
	tmpl := parseTemplate("letter.html")
	if tmpl == nil {
		log.Println("❌ letterHandler: шаблон не найден")
		http.Error(w, "Шаблон не найден", http.StatusInternalServerError)
		return
	}
	
	log.Println("📨 letterHandler: выполняем шаблон")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("❌ letterHandler: ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка рендеринга", http.StatusInternalServerError)
		return
	}
	log.Println("✅ letterHandler: страница отдана успешно")
}

func openLetterHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	user, ok := session.Values["user"].(User)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	data := struct {
		User User
	}{
		User: user,
	}
	
	tmpl := parseTemplate("letter_opened.html")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("❌ openLetterHandler: ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка рендеринга", http.StatusInternalServerError)
		return
	}
	log.Println("✅ openLetterHandler: страница отдана успешно")
}

func couponHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	http.Redirect(w, r, "/coupon-form", http.StatusSeeOther)
}

func couponFormHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	user := session.Values["user"].(User)
	
	data := struct {
		User User
	}{
		User: user,
	}
	
	tmpl := parseTemplate("coupon_form.html")
	tmpl.Execute(w, data)
}

func couponSubmitHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	r.ParseForm()
	
	coupon := CouponData{
		Date:      r.FormValue("date"),
		Time:      r.FormValue("time"),
		Place:     r.FormValue("place"),
		UserName:  session.Values["user"].(User).Name,
		CreatedAt: time.Now().Format("02.01.2006 15:04"),
	}
	
	session.Values["coupon"] = coupon
	session.Save(r, w)
	
	http.Redirect(w, r, "/coupon-receipt", http.StatusSeeOther)
}

func couponReceiptHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "quiz")
	if err != nil || session.Values["user"] == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	coupon, ok := session.Values["coupon"].(CouponData)
	if !ok {
		http.Redirect(w, r, "/coupon-form", http.StatusSeeOther)
		return
	}
	
	data := struct {
		Coupon CouponData
	}{
		Coupon: coupon,
	}
	
	tmpl := parseTemplate("coupon_receipt.html")
	tmpl.Execute(w, data)
}