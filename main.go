package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type DataBase struct {
	Data            map[string]*Expression
	NextId          int
	ExpressionsChan chan *Expression
	LastInputs      []*Expression
	OperationTime   *OperationTime
	SearchId        string
}

func (db *DataBase) Add(exp *Expression) {
	if err := exp.Parse(); err == nil {
		exp.Status = "in queue"
		db.ExpressionsChan <- exp
	}
	exp.Id = strconv.Itoa(db.NextId)
	db.NextId++
	db.Data[exp.Id] = exp
}

var (
	inputExpressionTemplate = template.Must(template.ParseFiles("static/templates/inputExpressionTemplate.html"))
	inputListTemplate       = template.Must(template.ParseFiles("static/templates/inputListTemplate.html"))
	listExpressionsTemplate = template.Must(template.ParseFiles("static/templates/listExpressionsTemplate.html"))
	configurationTemplate   = template.Must(template.ParseFiles("static/templates/configurationTemplate.html"))
	db                      = DataBase{
		Data:            make(map[string]*Expression),
		NextId:          1,
		ExpressionsChan: make(chan *Expression),
		OperationTime:   NewOperationTime(),
	}
)

func InputExpressionHandler(w http.ResponseWriter, r *http.Request) {
	inputExpressionTemplate.Execute(w, db.LastInputs)
}

func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	input := r.PostFormValue("expression")
	exp := &Expression{
		Input:        input,
		CreationTime: time.Now(),
	}
	db.Add(exp)
	db.LastInputs = append([]*Expression{exp}, db.LastInputs...)
	if len(db.LastInputs) >= 11 {
		db.LastInputs = db.LastInputs[:10]
	}
	inputListTemplate.Execute(w, db.LastInputs)
}

func ListExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	db.SearchId = r.URL.Query().Get("id")
	listExpressionsTemplate.Execute(w, db)
}

func FindExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	listExpressionsTemplate.Execute(w, db.Data)
}

func ConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	configurationTemplate.Execute(w, db.OperationTime)
}

func ChangeConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	timePlus, _ := strconv.Atoi(r.PostFormValue("time-plus"))
	timeMinus, _ := strconv.Atoi(r.PostFormValue("time-minus"))
	timeMultiply, _ := strconv.Atoi(r.PostFormValue("time-multiply"))
	timeDivide, _ := strconv.Atoi(r.PostFormValue("time-divide"))
	db.OperationTime = &OperationTime{
		TimePlus:     timePlus,
		TimeMinus:    timeMinus,
		TimeMultiply: timeMultiply,
		TimeDivide:   timeDivide,
	}
}

func init() {
	for range 10 {
		go func() {
			for exp := range db.ExpressionsChan {
				exp.Status = "calculating"
				exp.Calculate()
			}
		}()
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", InputExpressionHandler).Methods("GET")
	r.HandleFunc("/add-expression", AddExpressionHandler).Methods("POST")
	r.HandleFunc("/list-expressions", ListExpressionsHandler).Methods("GET")
	r.HandleFunc("/configuration", ConfigurationHandler).Methods("GET")
	r.HandleFunc("/configuration/change", ChangeConfigurationHandler).Methods("PUT")

	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fs))

	log.Println("Starting server at port :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
