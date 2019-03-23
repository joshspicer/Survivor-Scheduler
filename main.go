/*
	SURVIVOR SCHEDULER
	Josh Spicer <https://joshspicer.com/>
	2019 March 23
*/

package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"firebase.google.com/go"
	"fmt"
	"google.golang.org/api/option"
	"html/template"
	"log"
	"net/http"
)

/*
Inits firebase application with provided credential file
*/
func initializeAppWithServiceAccount() *firebase.App {
	opt := option.WithCredentialsFile("/Users/joshspicer/Documents/secrets/survivor/survivor-firebase.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	return app
}

/*
Inits firestore instance.
*/
func initFirestore() *firestore.Client {
	// Init app
	app := initializeAppWithServiceAccount()

	// Init firestore service
	client, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	return client
}

/*
Main function. Entry point of program.
*/
func main() {
	//client := initFirestore()

	templates := template.Must(template.ParseFiles("templates/welcome-template.html"))
	http.Handle("/static/", //final url can be anything
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		//Takes the name from the URL query e.g ?name=Martin, will set welcome.Name = Martin.
		if name := r.FormValue("name"); name != "" {
			welcome.Name = name
		}
		//If errors show an internal server error message
		//I also pass the welcome struct to the welcome-template.html file.
		if err := templates.ExecuteTemplate(w, "welcome-template.html", welcome); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	//_, _, err := client.Collection("users").Add(context.Background(), map[string]interface{}{
	//	"first": "Ada",
	//	"last":  "Lovelace",
	//	"born":  1815,
	//})
	//if err != nil {
	//	log.Fatalf("Failed adding alovelace: %v", err)
	//}

	fmt.Println("Listening")
	fmt.Println(http.ListenAndServe(":25500", nil))

}
