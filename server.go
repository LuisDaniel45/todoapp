package main

import (
	"net/http"
	"os"
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
        "register.html",
        "login.html",
    );
    if err != nil {panic(err) }


    session_manager := make(map[string]string) 
    users := make(map[string]string) 
    todo := make(map[string][]string);
    var user_counter int;

    http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            r.ParseForm();
            if r.PostForm.Has("username") && 
               r.PostForm.Has("password") {
                   user := r.PostForm["username"][0];
                   if _, ok := users[user]; ok ||  
                   user == ""  || r.PostForm["password"][0] == "" {
                       t.ExecuteTemplate(w, "error.html", err_values{400, 
                       "Bad Request, username already taken, or missing input"})
                       return
                   }

                   users[user] = r.PostForm["password"][0];
                   user_counter++;

                   uuid, err := os.ReadFile("/proc/sys/kernel/random/uuid")
                   if err != nil {
                       println("ERROR: generating uuid for session_id error")
                       t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
                       return
                   }

                   session_id := string(uuid[:len(uuid)-1])
                   session_manager[session_id] = user  
                   http.SetCookie(w, &http.Cookie{
                       Name: "auth", 
                       Value: session_id, 
                   });

                   http.Redirect(w, r, "/", 301);
                   return
            }
        }

        err = t.ExecuteTemplate(w, "register.html", nil);
        if err != nil {
            println("ERROR: register render");
        }
    });

    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            r.ParseForm()
            if r.PostForm.Has("username") || 
               r.PostForm.Has("password") {
                   user := r.PostForm["username"][0];
                   password := r.PostForm["password"][0];
                   if user == "" || password == ""  {
                       err = t.ExecuteTemplate(w, "error.html", err_values{401, 
                                        "Bad Request: missing username or password"})
                       if err != nil {
                           println("ERROR: render error.html for missing username or password");
                           println(err);
                       }
                       return

                   } else if p, ok := users[user]; !ok || p != password{  
                       err = t.ExecuteTemplate(w, "error.html", err_values{401, 
                                        "Bad Request: invalid username or password"})
                       if err != nil {
                           println("ERROR: render error.html for invalid username or password");
                           println(err);
                       }
                       return
                   }

                   uuid, err := os.ReadFile("/proc/sys/kernel/random/uuid")
                   if err != nil {
                       println("ERROR: generating uuid for session_id error")
                       t.ExecuteTemplate(w, "error.html", err_values{501, "Internal Server Error"})
                       return
                   }

                   session_id := string(uuid[:len(uuid)-1])
                   session_manager[session_id] = user  
                   http.SetCookie(w, &http.Cookie{
                       Name: "auth", 
                       Value: session_id, 
                   });

                   http.Redirect(w, r, "/", 301)
                   return
            }
        }

        err = t.ExecuteTemplate(w, "login.html", nil);
        if err != nil {
            println("ERROR: login render");
        }
    });

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            w.WriteHeader(404)
            err = t.ExecuteTemplate(w, "error.html", err_values{404, "Not Found"})
            if err != nil {
                println("ERROR: render error.html");
            }
            return;
        }

        cookie, err := r.Cookie("auth");
        if err != nil {
            w.WriteHeader(403)
            err = t.ExecuteTemplate(w, "error.html", err_values{403, "Forbidden Acces, Not Autherize"})
            if err != nil {
                println("ERROR: render errror.html");
            }
            return 
        }

        user, ok := session_manager[cookie.Value];
        if !ok {
            w.WriteHeader(403)
            err = t.ExecuteTemplate(w, "error.html", err_values{403, "Forbidden Access, Not Autherize"})
            if err != nil {
                println("ERROR: render errror.html");
            }
            return
        }


        if r.Method == "POST" {
            r.ParseForm();
            if r.PostForm.Has("todo") {
                todo[user] = append(todo[user], r.PostForm["todo"][0])
            }
        }

        err = t.ExecuteTemplate(w, "main.html", todo[user])
        if err != nil {
            println("ERROR: content render");
        }
    });

    err = http.ListenAndServe(":8080", nil);
    if err != nil {
        println("ERROR: opening port");
    }
}
