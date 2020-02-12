package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tohirov1994/database"
	"github.com/tohirov1994/managers-core/pkg/core"
	"github.com/tohirov1994/terminal"
	"log"
	"os"
	"strconv"
	"strings"
)

var UserName string

func main() {
	log.Print("start application")
	log.Print("open db")
	dbDriver := "sqlite3"
	dataSource := "../database/db.sqlite"
	db, err := sql.Open(dbDriver, dataSource)
	if err != nil {
		log.Fatalf("can't open db: %v", err)
	}
	defer func() {
		log.Print("close db")
		if err := db.Close(); err != nil {
			log.Fatalf("can't close db: %v", err)
		}
	}()
	err = core.Init(db)
	if err != nil {
		log.Fatalf("can't init db: %v", err)
	}
	terminal.Cleaner()
	fmt.Println("Welcome!!!")
	log.Print("start operations loop")
	operationsLoop(db, beforeAuth, unauthorizedOperationsLoop)
	log.Print("finish operations loop")
	log.Print("finish application")
}

func operationsLoop(db *sql.DB, commands string, loop func(db *sql.DB, cmd string) bool) {
	for {
		fmt.Println(commands)
		var cmd string
		_, err := fmt.Scan(&cmd)
		if err != nil {
			log.Fatalf("Can't read input: %v", err)
		}
		if exit := loop(db, strings.TrimSpace(cmd)); exit {
			return
		}
	}
}

func handleLogin(db *sql.DB) (okLogin bool, errLogin error) {
	fmt.Println("Please enter your authentications")
	fmt.Print("Login: ")
	_, errLogin = fmt.Scan(&UserName)
	if errLogin != nil {
		return false, errLogin
	}
	var password string
	fmt.Print("Password: ")
	_, errLogin = fmt.Scan(&password)
	if errLogin != nil {
		return false, errLogin
	}
	okLogin, errLogin = core.SignIn(UserName, password, db)
	if errLogin != nil {
		return false, errLogin
	}
	return okLogin, errLogin
}

func unauthorizedOperationsLoop(db *sql.DB, cmd string) (exit bool) {
	switch cmd {
	case "1":
		ok, err := handleLogin(db)
		if err != nil {
			log.Printf("can't handle login: %v", err)
			return true
		}
		if !ok {
			terminal.Cleaner()
			fmt.Println("Wrong Login or Password. Please try again.")
			return false
		}
		fmt.Printf("\nWelcome: %s\n", UserName)
		operationsLoop(db, afterAuth, authorizedOperationsLoop)
	case "2":
		fmt.Printf("Available ATMs: \n")
		ATMs, err := core.ATMsGet(db)
		if err != nil {
			log.Printf("can't get the ATMs data: %v", err)
			return true
		}
		atmPrint(ATMs)
		fmt.Printf("\n")
		return false
	case "q":
		return true
	default:
		fmt.Printf("You entered the wrong command: %s\n", cmd)
	}
	return false
}

func authorizedOperationsLoop(db *sql.DB, cmd string) (exit bool) {
	switch cmd {
	case "1":
		err := addClient(db)
		if err != nil {
			log.Printf("can't add the new client: %v", err)
			return true
		}
	case "2":
		err := addCard(db)
		if err != nil {
			log.Printf("can't add card to client: %v", err)
			return true
		}
	case "3":
		err := addService(db)
		if err != nil {
			log.Printf("can't add the service: %v", err)
			return true
		}
	case "4":
		err := addAtm(db)
		if err != nil {
			log.Printf("can't add ATM: %v", err)
			return true
		}
	case "5":
		fmt.Println("Start export")
		Result, err := handleExport(db)
		if err != nil {
			log.Printf("can't Export data: %v", err)
		}
		if Result != "Successfully!" {
			fmt.Println("Export failed")
		}
		log.Println("Export was successfully")
	case "6":
		fmt.Println("We can not import data")
	case "q":
		return true
	default:
		fmt.Printf("You entered the wrong command: %s\n", cmd)
	}
	return false
}

func atmPrint(ATMs [] core.ATMStruct) {
	for _, atm := range ATMs {
		fmt.Printf(
			"id: %d, City: %s, District: %s, Street: %s\n",
			atm.Id, atm.City, atm.District, atm.Street,
		)
	}
}

func addClient(db *sql.DB) (errClient error) {
	fmt.Println("Enter the Client data")
	var name string
	fmt.Print("Name: ")
	_, errClient = fmt.Scan(&name)
	if errClient != nil {
		return errClient
	}
	var surname string
	fmt.Print("Surname: ")
	_, errClient = fmt.Scan(&surname)
	if errClient != nil {
		return errClient
	}
	var loginC string
	fmt.Print("Login(length more 3): ")
	_, errClient = fmt.Scan(&loginC)
	if errClient != nil {
		return errClient
	}
	if len(loginC) <= 3 {
		fmt.Printf("Enter the login is length 4 or lengter")
		os.Exit(0)
	}
	var passC string
	fmt.Print("Password(length more 3): ")
	_, errClient = fmt.Scan(&passC)
	if errClient != nil {
		return errClient
	}
	if len(passC) <= 3 {
		fmt.Printf("Enter the password is length 4 or lengter")
		os.Exit(0)
	}
	errClient = core.AddClient(name, surname, loginC, passC, db)
	if errClient != nil {
		return
	}
	fmt.Println("The client was added success!")
	return nil
}

func addCard(db *sql.DB) (errCard error) {
	fmt.Println("Enter the card data:")
	lastPAN, errCard := core.PANLastPlusOne(db)
	if errCard != nil {
		fmt.Print("Error generates PAN for card: ")
		os.Exit(0)
	}
	fmt.Printf("The card number generated: %d\n", lastPAN)
	var clientId int64
	fmt.Print("Enter client ID: ")
	_, errCard = fmt.Scan(&clientId)
	if errCard != nil {
		return errCard
	}
	acceptId, errCard := core.CheckIdClient(clientId, db)
	if errCard != nil {
		fmt.Printf("I can't check Id client: %v\n", errCard)
		os.Exit(0)
	}
	clientId = acceptId
	Name, Surname, errCard := core.GetNameSurnameFromIdClient(clientId, db)
	if errCard != nil {
		fmt.Printf("I can't get the client Name and Surname from this idClient: %d i have error: %v\n", clientId, errCard)
		os.Exit(0)
	}
	fmt.Printf("This name: %s, and surname: %s will be used.\n", Name, Surname)
	NameSurname := strings.ToUpper(strings.Join([]string{Name, Surname}, " "))
	fmt.Print("Enter PIN card(length must be 4): ")
	var tmpPIN string
	_, errCard = fmt.Scan(&tmpPIN)
	if errCard != nil {
		return
	}
	var lengthPIN int64
	lengthPIN = int64(len(tmpPIN))
	if lengthPIN != 4 {
		fmt.Printf("PIN card length must be 4")
		os.Exit(0)
	}
	var pinCard int64
	pinCard, _ = strconv.ParseInt(tmpPIN, 10, 64)
	fmt.Printf("Enter CVV card(length must be 3): ")
	var tmpCVV string
	_, errCard = fmt.Scan(&tmpCVV)
	if errCard != nil {
		return
	}
	var lengthCVV int64
	lengthCVV = int64(len(tmpCVV))
	if lengthCVV != 3 {
		fmt.Printf("Check CVV! and try again.\n")
		os.Exit(0)
	}
	var CVVCard int64
	CVVCard, _ = strconv.ParseInt(tmpCVV, 10, 64)
	fmt.Print("Enter Balance card zero or more: ")
	var balanceAmount int64
	_, errCard = fmt.Scan(&balanceAmount)
	if errCard != nil {
		return errCard
	}
	if balanceAmount < 0 {
		fmt.Printf("Balance card can't be lower than zero!\n")
		os.Exit(0)
	}
	var validityCard int64
	fmt.Print("Enter Validity card 1222: ")
	_, errCard = fmt.Scan(&validityCard)
	if errCard != nil {
		return errCard
	}
	errCard = core.AddCardToClient(lastPAN, pinCard, balanceAmount, NameSurname, CVVCard, validityCard, clientId, db)
	if errCard != nil {
		return errCard
	}
	fmt.Printf("Card was successfully added to client by name: %s, by surname: %s by identifier: %d!\n\n", Name, Surname, clientId)
	return nil
}

func addService(db *sql.DB) (errService error) {
	fmt.Println("Enter the service data")
	var serviceName string
	fmt.Print("Service Name: ")
	_, errService = fmt.Scan(&serviceName)
	if errService != nil {
		return errService
	}
	errService = core.AddServiceToTheBank(serviceName, db)
	if errService != nil {
		return errService
	}
	fmt.Printf("The service: %s was added success!\n", serviceName)
	return nil
}

func addAtm(db *sql.DB) (errAtm error) {
	fmt.Println("Enter the ATM data:")
	var cityName string
	fmt.Print("City: ")
	city := bufio.NewReader(os.Stdin)
	cityName, errAtm = city.ReadString('\n')
	if errAtm != nil {
		return errAtm
	}
	var districtName string
	fmt.Print("District: ")
	district := bufio.NewReader(os.Stdin)
	districtName, errAtm = district.ReadString('\n')
	if errAtm != nil {
		return errAtm
	}
	fmt.Print("Street: ")
	var streetName string
	street := bufio.NewReader(os.Stdin)
	streetName, errAtm = street.ReadString('\n')
	if errAtm != nil {
		return errAtm
	}
	errAtm = core.AddAtmToTheBank(cityName, districtName, streetName, db)
	if errAtm != nil {
		return errAtm
	}
	fmt.Printf("The ATM was success added to city: %s, district: %s, street: %s!\n", cityName, districtName, streetName)
	return nil
}

func handleExport(db *sql.DB) (Result string, errExp error) {
	Result, errExp = core.DoAllForMe(db)
	if errExp != nil {
		return "I can't Exporting data", errExp
	}
	if Result != "YOU ARE LUCKY =)" {
		fmt.Printf("You are not lucky =(")
	}
	return "Successfully!", nil
}
