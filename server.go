package main

import (
	"net/http"
	"text/template"
)

type err_values struct { 
    Code int 
    Msg string
}

func main()  {
    t, err := template.ParseFiles(
        "index.html",
        "main.html",
        "error.html",
    );
    if err != nil {panic(err) }


    var counter int;
    var str []string;
    http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
    });

    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

    });

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            w.WriteHeader(404)
            err = t.ExecuteTemplate(w, "error.html", err_values{404, "Not Found"})
            if err != nil {
                println("ERROR: content render");
            }
            return;
        }

        if r.Method == "POST" {
            r.ParseForm();
            if r.PostForm.Has("todo") {
                str = append(str, r.PostForm["todo"][0])
                counter++;
            }
        }

        err = t.ExecuteTemplate(w, "main.html", str)
        if err != nil {
            println("ERROR: content render");
        }
    });

    err = http.ListenAndServe(":8080", nil);
    if err != nil {
        println("ERROR: opening port");
    }
}
