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
    const button = document.getElementById("button")
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

function todo_func(element) {
    const button = document.querySelector("input[type='submit']");
    if (element.value != "") {
        button.disabled = false;
        return;
    }
    button.disabled = true;
}

function todo_done(element) {
    fetch("/delete_task?task=" + element.value)
        .then((response) => {
            if (response.status == 200) {
                element.parentElement.remove();
                console.log(response.body);
                return
            }
            console.log("ERROR: Somthing Went Wrong");
        })
}
