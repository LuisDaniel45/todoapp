package main

import (
	"net/http"
	"text/template"
)

type err_values struct { 
    Code int 
    Msg string
}
type html_values struct { 
    Type string 
    Value any 
}

func main()  {
    t, err := template.ParseFiles(
        "index.html",
        "main.html",
        "error.html",
    )
    if err != nil {panic(err) }

    var counter int;
    var str []string;

    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

    });

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            w.WriteHeader(404)
            err = t.Execute(w, html_values{"error", err_values{404, "Not Found"}})
            if err != nil {
                println("ERROR: executing 'err_html'");
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

        err = t.Execute(w, html_values{"[]string", str});
        if err != nil {
            println("ERROR: executing 'index'");
        }
    });

    http.ListenAndServe(":8080", nil);
}
