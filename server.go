package main

import (
	"fmt"
	"net/http"
	"os"
	"text/template"
)
type err_values struct { 
    Code int 
    Msg string
}

func main()  {
    index, err := get_template("index.html", "index");
    if err != nil {panic(err) }

    err_html, err := get_template("error.html", "error");
    if err != nil {panic(err) }

    var counter int;
    var str []string;

    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

    });

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            w.WriteHeader(404)
            err = err_html.Execute(w, err_values{404, "Not Found"})
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

        err = index.Execute(w, str);
        if err != nil {
            println("ERROR: executing 'index'");
        }
    });

    http.ListenAndServe(":8080", nil);
}


func get_template(filename string, temp string) (*template.Template, error) { 
    content, err := os.ReadFile(filename);
    if err != nil { 
        fmt.Println("Error: Opening file");
        return nil, err
    }

    tmpl, err := template.New(temp).Parse(string(content));
    if err != nil { 
        fmt.Println("Error: Creating template");
        return nil, err
    }

    return tmpl, nil 
}
