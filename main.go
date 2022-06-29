package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	id         = "id"
	item       = "item"
	fileName   = "fileName"
	operation  = "operation"
	addOp      = "add"
	findByIdOp = "findById"
	removeOp   = "remove"
	listOp     = "list"
)

type Arguments map[string]string
type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   uint   `json:"age"`
}

func parseArgs() Arguments {
	flagOperation := flag.String(operation, "", "Allowed values: [add|findById|remove|list]")
	flagFileName := flag.String(fileName, "", "Path to the JSON file with user's data.")
	flagItem := flag.String(item, "", "User JSON, for example {''id'': ''1'', ''email'': ''email@test.com'', ''age'': 23}")
	flagId := flag.String(id, "", "User Identifier, should be greater then zero")
	flag.Parse()

	return Arguments{
		operation: *flagOperation,
		item:      *flagItem,
		id:        *flagId,
		fileName:  *flagFileName}
}

func Perform(args Arguments, writer io.Writer) error {
	operationArg := args[operation]
	if len(operationArg) == 0 {
		return errors.New("-operation flag has to be specified")
	}
	fileNameArg := args[fileName]
	if len(fileNameArg) == 0 {
		return errors.New("-fileName flag has to be specified")
	}
	idArg := args[id]
	if (operationArg == removeOp || operationArg == findByIdOp) && len(idArg) == 0 {
		return errors.New("-id flag has to be specified")
	}
	itemArg := args[item]
	if (operationArg == addOp) && len(itemArg) == 0 {
		return errors.New("-item flag has to be specified")
	}
	switch operationArg {
	case addOp:
		return addUser(itemArg, fileNameArg, writer)
	case findByIdOp:
		return findUserById(idArg, fileNameArg, writer)
	case removeOp:
		return removeUser(idArg, fileNameArg, writer)
	case listOp:
		return listUsers(fileNameArg, writer)
	default:
		return fmt.Errorf("Operation %s not allowed!", operationArg)
	}
	return nil
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func removeUser(userId, fileName string, writer io.Writer) error {
	users, err := loadUsersFromFile(fileName)
	if err != nil {
		return err
	}
	found := false
	for i, cUser := range users {
		if cUser.Id == userId {
			found = true
			users = append(users[:i], users[i+1:]...)
		}
	}
	if !found {
		return fmt.Errorf("User with id %s was not found", userId)
	}
	err = saveUsersToFile(users, fileName)
	if err != nil {
		return err
	}
	return nil
}

func listUsers(fileName string, writer io.Writer) error {
	users, err := loadUsersFromFile(fileName)
	if err != nil {
		return err
	}
	usersData, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("Error while marshaling users to json file: %w", err)
	}
	writer.Write(usersData)
	return nil
}

func findUserById(idArg, fileName string, writer io.Writer) error {
	users, err := loadUsersFromFile(fileName)
	if err != nil {
		return err
	}
	user := User{Id: "", Email: "", Age: 0}
	for _, cUser := range users {
		if cUser.Id == idArg {
			user = cUser
		}
	}
	if user.Id == "" {
		writer.Write([]byte(""))
		return fmt.Errorf("User with id %s was not found", idArg)
	}
	userData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("Error while marshaling users to json file: %w", err)
	}
	writer.Write(userData)
	return nil
}

func addUser(item, fileName string, writer io.Writer) error {
	var pendingUser User
	err := json.Unmarshal([]byte(item), &pendingUser)
	if err != nil {
		return fmt.Errorf("Error to unmarshal a user defined with JSON: %w", err)
	}
	users, err := loadUsersFromFile(fileName)
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Id == pendingUser.Id {
			writer.Write([]byte("User with id " + user.Id + " already exists"))
			return nil
		}
	}
	users = append(users, pendingUser)
	err = saveUsersToFile(users, fileName)
	if err != nil {
		return fmt.Errorf("failed to save users: %w", err)
	}
	return nil
}

func loadUsersFromFile(fileName string) ([]User, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("Error while opening file with users: %w", err)
	}
	defer file.Close()

	usersData, err := io.ReadAll(file)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("Error while reading users from file: %w", err)
	}
	var users []User
	if len(usersData) > 0 {
		err = json.Unmarshal(usersData, &users)
		if err != nil {
			return nil, fmt.Errorf("Error while unmarshaling users from json file: %w", err)
		}
	}
	return users, nil
}

func saveUsersToFile(users []User, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("Error while opening file with users: %w", err)
	}
	defer file.Close()

	jsonData, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("Error while marshaling users to json file: %w", err)
	}
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("Error while writing users to a file: %w", err)
	}

	return nil
}
