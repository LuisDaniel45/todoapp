function drag(ev) {
    ev.dataTransfer.setData("text", ev.target.id)
}

function drop(ev) {
    ev.preventDefault();
    const data = ev.dataTransfer.getData("text");
    const list = document.querySelector(".todo > div"); 


    const target = (ev.target.id === "") ? ev.target.parentElement: ev.target;
    const elem = document.getElementById(data);
    fetch("/change_priority?task=" + elem.querySelector('button').value +
          "&priority=" + target.id + "&direction=" + ((data > target.id)? "up": "down"))
        .then(res => res.text())
        .then(bod => console.log(bod));

    elem.id = target.id;
    if (data < target.id)  {
        for (let i = elem.nextElementSibling; i != target.nextElementSibling; i = i.nextElementSibling) {
            i.id--;
        }
        list.insertBefore(elem, target.nextElementSibling);
        return;
    }

    for (let i = target; i != elem; i = i.nextElementSibling) {
        i.id++;
    }
    list.insertBefore(elem, target);
}

const confirm = document.getElementById("confirm")
if (confirm != null) {
    confirm.addEventListener("input", register)
}

ev = false;
function register() {
    if (!ev) {
        document.getElementById("password").addEventListener("input", register)
        ev = true
    }
    const conf = confirm.value
    const pass =  document.getElementById("password").value
    const msg = document.getElementById("msg")
    const button = document.querySelector("#register > input[type='submit']")
    if (conf == "" || pass == "") {
        msg.innerText = "Missing password or confirmation"
        button.disabled = true
        return
    }
    else if (conf != pass) {
        msg.innerText = "Password and confirmation does not match"
        button.disabled = true
        return
    }
    msg.innerText = ""
    button.disabled = false 
}

const form = document.querySelector(".todo > form")
if (form != null) {
    form.addEventListener("submit", (ev) => {
        ev.preventDefault()
        const input = form.querySelector("input[type='text'");
        const data = input.value;
        const key = input.name;
        input.value = ""
        fetch("/", {
            method: "POST",
            mode: "cors",
            headers: {
                "Content-Type": "application/x-www-form-urlencoded"
            },
            body: key + "=" + data
        })
        .then((response) =>  {
            if (response.ok) {
                return response.text()
            }
        })
        .then(task_index => {
            const todo_list = form.parentElement.querySelector("div");
            const container = document.createElement("div")

            const button = document.createElement("button");
            button.type = "button";
            button.value= task_index;
            button.innerHTML = "done?";
            button.onclick = () => {todo_done(button)};

            const text = document.createElement("p");
            text.innerHTML = data;

            container.appendChild(button);
            container.appendChild(text);
            todo_list.appendChild(container);
        });
    });
}

function todo_func(element) {
    const button = document.querySelector("input[type='submit']");
    if (element.value != "") {
        button.disabled = false; return;
    }
    button.disabled = true;
}

function todo_done(element) {
    fetch("/delete_task?task=" + element.value)
        .then((response) => {
            if (response.status == 200) {
                element.parentElement.remove();
                return
            }
            console.log("ERROR: Somthing Went Wrong");
        })
}
