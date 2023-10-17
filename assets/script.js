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
};
