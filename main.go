package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

//Расписание
type Schedule struct {
	ArrivalDate   string
	DepartureDate string
	RouteNumber   int
	IdStation     int
	NumberOfTrain int
	IdRoute       int
	Price         int
}

//Расписание от станции до станции
type DSchedule struct {
	ArrivalDate    string
	DepartureDate  string
	RouteNumber    int
	IdStation      int
	NumberOfTrain  int
	IdRoute        int
	ArrivalDate2   string
	DepartureDate2 string
	RouteNumber2   int
	IdStation2     int
	NumberOfTrain2 int
	IdRoute2       int
	Price          int
}

type TrainStation struct {
	IdStation int
}

type Passenger struct {
	idPassenger int
	name        string
	patronymic  string
	surname     string
	passport    string
	password    string
}

var database *sql.DB
var store = sessions.NewCookieStore([]byte("something-very-secret"))

func MyHandler(w http.ResponseWriter, r *http.Request, id int) {
	session, _ := store.Get(r, "session-name")
	session.Values["id"] = id
	err := session.Save(r, w)
	log.Println("Почти отработал функцию myHandler и идёт дальше")
	if err != nil {
		log.Println("Попал в  ошибку на конце myHandler")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

//Формирование билета
func PushTicketIntoDB(w http.ResponseWriter, r *http.Request, idStation int, idStation2 int, RouteNumber int) {

	place := 0
	carr := 0

	session, err2 := store.Get(r, "session-name")
	log.Println("session ID: ", session.Values["id"])
	if err2 != nil {
		log.Println(err2)

	}

	var tmp sql.NullString
	var tmp1 sql.NullString
	var tmp2 sql.NullString
	var tmp3 sql.NullString
	var err0 error
	err5000 := database.QueryRow("Select №_поезда from Станции_поезда where idСтанция = ? and idМаршрут = ?", idStation, RouteNumber).Scan(&tmp1)
	if err5000 != nil {
		log.Println(err5000)
	}
	var resNumberOfTrain string
	resNumberOfTrain = tmp1.String
	NumberOfTrain, err5001 := strconv.Atoi(resNumberOfTrain)
	if err5001 != nil {
		log.Println(err5001)
	}
	err5002 := database.QueryRow("Select Время_отправления from Станции_поезда where idСтанция = ? and idМаршрут = ?", idStation, RouteNumber).Scan(&tmp2)
	if err5002 != nil {
		log.Println(err5002)
	}

	DepartureDate := tmp2.String
	//Смотрим сколько вагонов в нужном поезде
	err0 = database.QueryRow("select max(№_Вагона) from trains.Вагон where №_поезда = ?", NumberOfTrain).Scan(&tmp)
	if err0 != nil {
		log.Println(err0)
	}
	//Берем цену билета
	err5005 := database.QueryRow("Select Тип_поезда.стоимость from Станции_поезда join Поезд on Станции_поезда.№_поезда = Поезд.№_Поезда join Тип_поезда on Поезд.Тип_поезда = Тип_поезда.Тип_поезда where Станции_поезда.idСтанция = ? and Станции_поезда.idМаршрут = ?", idStation, RouteNumber).Scan(&tmp3)
	if err5005 != nil {
		log.Println(err5005)
	}
	PriceStr := tmp3.String
	Price, err5006 := strconv.Atoi(PriceStr)
	if err5006 != nil {
		log.Println(err5006)
	}
	log.Println("Price before: ", Price)
	var result string
	result = tmp.String
	resultInt, err := strconv.Atoi(result)
	if err != nil {
		log.Println(err)
	}
	//Проверка на существование места в выбранном поезде
	for i := 1; i <= resultInt; i++ {
		var tmp1 sql.NullString
		var err10 error
		err10 = database.QueryRow("select max(№_Места) from trains.Билет where №_Поезда = ? and №_Вагона = ?", NumberOfTrain, i).Scan(&tmp1)
		//	rows, err := database.Query("select max(№_Места) from trains.Билет where №_Поезда = ? and №_Вагона = ?",
		//number, i)
		err := database.QueryRow("Select Тип_поезда.стоимость from Станции_поезда join Поезд on Станции_поезда.№_поезда = Поезд.№_Поезда join Тип_поезда on Поезд.Тип_поезда = Тип_поезда.Тип_поезда  where Станции_поезда.idСтанция = ? and Станции_поезда.idМаршрут = ?", idStation, NumberOfTrain)
		if err10 != nil {
			log.Println(err10)
		}
		var result2 string
		result2 = tmp1.String
		resultInt2, err5004 := strconv.Atoi(result2)

		if err5004 != nil {
			log.Println(err)
		}
		if resultInt2 < 6 {
			place = resultInt2 + 1
			carr = i
			break
		}
		if i == resultInt && place == 0 {
			//io.WriteString(w, "error")
			log.Print("В поезде не осталось мест, извините.")
			return
		}
	}

	//Отправка информации о новом билете в базу данных
	fmt.Println(carr)
	fmt.Println(place)
	log.Println("Price after: ", Price)
	_, err = database.Exec("insert into trains.Билет (стоимость,Дата_отправления ,idПассажир,idСтанция_1, idСтанция_2, idКассир, №_Места, №_Вагона,№_Поезда) values (?,?,?,?,?,?,?,?,?)",
		Price, DepartureDate, session.Values["id"], idStation, idStation2,
		2, place, carr, NumberOfTrain)
	if err != nil {
		log.Println(err)
	}
	//io.WriteString(w, "success")

}

//------------------------Фильтр_расписания-----------------------------
func Filter(w http.ResponseWriter, r *http.Request) {
	/*ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)*/
	var b []byte
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		nameOfFirstStation := r.FormValue("nameOfFirstStation")
		nameOfSecondStation := r.FormValue("nameOfSecondStation")
		if nameOfFirstStation == nameOfSecondStation {
			log.Println("Станция отбытия и прибытия не должны совпадать!")
		}
		rows, err := database.Query("select idСтанция from trains.Станция where  Название = ? ",
			nameOfFirstStation)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		stations := []TrainStation{}
		for rows.Next() {
			s := TrainStation{}
			err := rows.Scan(&s.IdStation)
			if err != nil {
				log.Println(err)
				continue
			}
			stations = append(stations, s)
		}
		//rows2,err := database.Query("Select Станции_поезда.*, Тип_поезда.стоимость from Станции_поезда join Поезд on Станции_поезда.№_поезда = Поезд.№_Поезда join Тип_поезда on Поезд.Тип_поезда = Тип_поезда.Тип_поезда")
		//////////
		rows3, err := database.Query("select idСтанция from trains.Станция where  Название = ? ",
			nameOfSecondStation)
		if err != nil {
			log.Println(err)
		}
		defer rows3.Close()
		for rows3.Next() {
			s := TrainStation{}
			err := rows3.Scan(&s.IdStation)
			if err != nil {
				log.Println(err)
				continue
			}
			stations = append(stations, s)
		}
		/////
		//fmt.Println(stations)
		log.Println("Информация о станциях: ", stations)
		rows2, err := database.Query("Select Станции_поезда.*, Тип_поезда.стоимость from Станции_поезда join Поезд on Станции_поезда.№_поезда = Поезд.№_Поезда join Тип_поезда on Поезд.Тип_поезда = Тип_поезда.Тип_поезда where idСтанция = ? or idСтанция = ?",
			stations[0].IdStation, stations[1].IdStation)
		//	rows2, err := database.Query("Select * from trains.Станции_поезда where idСтанция = ? or idСтанция = ?",
		//	stations[0].IdStation, stations[1].IdStation)
		//	fmt.Println(stations[0].IdStation, stations[1].IdStation)
		if err != nil {
			log.Println(err)
		}
		order := stations[0].IdStation < stations[1].IdStation
		//fmt.Print(order)
		defer rows2.Close()
		route := []Schedule{}
		for rows2.Next() {
			r := Schedule{}
			err := rows2.Scan(&r.ArrivalDate, &r.DepartureDate, &r.RouteNumber, &r.IdStation, &r.NumberOfTrain, &r.IdRoute, &r.Price)
			if err != nil {
				log.Println(err)
				continue
			}
			route = append(route, r)
		}
		//Выбор подходящих пассажиру маршрутов, учитывая направление движения
		var RouteList []DSchedule
		if order == true {

			for i := 0; i < len(route); i++ {
				for j := 0; j < len(route); j++ {
					if (route[i].IdStation < route[j].IdStation) && (route[i].IdRoute == route[j].IdRoute) {
						GoodRoute := DSchedule{route[i].ArrivalDate, route[i].DepartureDate,
							route[i].RouteNumber, route[i].IdStation, route[i].NumberOfTrain,
							route[i].IdRoute, route[j].ArrivalDate, route[j].DepartureDate, route[j].RouteNumber,
							route[j].IdStation, route[j].NumberOfTrain, route[j].IdRoute, route[i].Price}

						if GoodRoute.IdStation == stations[0].IdStation &&
							GoodRoute.IdStation2 == stations[1].IdStation &&
							GoodRoute.IdStation < GoodRoute.IdStation2 &&
							GoodRoute.RouteNumber < GoodRoute.RouteNumber2 {
							RouteList = append(RouteList, GoodRoute)
						}
					}
				}
			}
		} else {
			for i := 0; i < len(route); i++ {
				for j := 0; j < len(route); j++ {
					if (route[i].IdStation > route[j].IdStation) && (route[i].IdRoute == route[j].IdRoute) {

						GoodRoute := DSchedule{route[i].ArrivalDate, route[i].DepartureDate,
							route[i].RouteNumber, route[i].IdStation, route[i].NumberOfTrain,
							route[i].IdRoute, route[j].ArrivalDate, route[j].DepartureDate, route[j].RouteNumber,
							route[j].IdStation, route[j].NumberOfTrain, route[j].IdRoute, route[i].Price}
						if GoodRoute.IdStation == stations[0].IdStation &&
							GoodRoute.IdStation2 == stations[1].IdStation &&
							GoodRoute.IdStation > GoodRoute.IdStation2 &&
							GoodRoute.RouteNumber < GoodRoute.RouteNumber2 {
							RouteList = append(RouteList, GoodRoute)
						}
					}
				}
			}

		}
		m := RouteList
		////                                    Передача подходящих расписаний в JSON
		b, err = json.Marshal(m)
		if err != nil {
			log.Println(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(b)

	}
}
func BestRouterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		idStationStr := r.FormValue("idStation")
		idStation2Str := r.FormValue("idStation2")
		RouteNumberStr := r.FormValue("IdRoute")
		idStation, err := strconv.Atoi(idStationStr)
		if err != nil {
			log.Println(err)
		}
		idStation2, err := strconv.Atoi(idStation2Str)
		if err != nil {
			log.Println(err)
		}
		RouteNumber, err := strconv.Atoi(RouteNumberStr)
		if err != nil {
			log.Println(err)
		}

		if err != nil {
			log.Println(err)
		}
		PushTicketIntoDB(w, r, idStation, idStation2, RouteNumber)
		io.WriteString(w, "success")
	}
}

//------------------------Вход на сайт------------------------------
func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		passport := r.FormValue("passport")
		password := r.FormValue("password")
		sum := sha256.Sum256([]byte(password))
		hash := fmt.Sprintf("%x", sum)

		//Смотрим что такой логин вообще есть
		rows, err := database.Query("select  count(паспорт) > 0  from trains.Пассажир where  паспорт = ?",
			passport)
		if err != nil {
			log.Println(err)
			io.WriteString(w, "error")
		}
		for rows.Next() {
			var result string
			err := rows.Scan(&result)
			if err != nil {
				log.Println(err)
				io.WriteString(w, "error")
			}

			if result == "1" {
				//если есть, то проверяем пароль
				rows, err := database.Query("select * from trains.Пассажир where паспорт = ?",
					passport)
				if err != nil {
					log.Println(err)
					io.WriteString(w, "error")
				}
				for rows.Next() {
					p := Passenger{}
					err := rows.Scan(&p.idPassenger, &p.name, &p.patronymic, &p.surname, &p.passport, &p.password)
					if err != nil {
						fmt.Println(err)
						io.WriteString(w, "error")
					}
					if p.password != hash {
						fmt.Println("Введен неверный логин или пароль")
						io.WriteString(w, "error")
					} else {
						fmt.Println("Добро пожаловать")
						id := p.idPassenger
						log.Println("Succes отправил и пошёл дальше, id отдал такой: ", id)
						MyHandler(w, r, id)
						io.WriteString(w, "success")
					}
				}
			} else {
				io.WriteString(w, "error")
				fmt.Println("Введен неверный логин или пароль")

			}
		}
		//io.WriteString(w, "success")
	}

}

//----------------Новый пользователь на сайте----------------------
func CreateNewUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		name := r.FormValue("name")
		patronymic := r.FormValue("patronymic")
		surname := r.FormValue("surname")
		passport := r.FormValue("passport")
		password := r.FormValue("password")

		//Заносим нового пользователя в базу данных
		_, err = database.Exec("INSERT INTO trains.Пассажир (имя, отчество, фамилия, паспорт, пароль) VALUES (?,?,?,?,SHA2(?,256))",
			name, patronymic, surname, passport, password)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Пользователь с такими паспортными данными уже существует!")
			io.WriteString(w, "error")
		} else {
			io.WriteString(w, "success")
		}
	}
}
func ListOfStations(w http.ResponseWriter, r *http.Request) {
	rows, err := database.Query("select Назавание from trains.Станция")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		err := rows.Scan(&n)
		if err != nil {
			log.Println(err)
		}
		names = append(names, n)
	}
	m := names
	b, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func LogOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Println(err)
	}
	session.Options.MaxAge = -1
	fmt.Println(session.Values["id"])
	//http.Redirect(w, r, "/", 302)
	io.WriteString(w, "success")
}
func main() {

	go func() {
		timeOfSleep := time.Now().
			AddDate(0, 0, 1).
			Round(time.Hour * 24).
			Sub(time.Now())

		<-time.After(timeOfSleep) // ждем начала новых суток
		_, err := database.Exec("TRUNCATE TABLE trains.Билет")
		if err != nil {
			log.Println(err)
		}
	}()

	db, err := sql.Open("mysql", "trains:oS24Kl0x@tcp(206.81.28.231)/trains")
	if err != nil {
		log.Print(err)
	}
	database = db
	defer db.Close()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	http.HandleFunc("/insert", CreateNewUserHandler)
	http.HandleFunc("/login", LoginUserHandler)
	http.HandleFunc("/ticket", Filter)
	http.HandleFunc("/buy", BestRouterHandler)
	http.HandleFunc("/logout", LogOut)
	fmt.Println("Server is listening...")
	http.ListenAndServe(":80", nil)

}
