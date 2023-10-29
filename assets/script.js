const drag_items = document.querySelectorAll(".draggable")
const con = document.querySelector(".todo > div");
for (let i = 0; i < drag_items.length; i++) {
    drag(drag_items[i], con, drag_items);
}

function transition(obj, y, speed) {
    const ani = obj.getAnimations()
    if (ani.length > 0) {
        const style = getComputedStyle(obj)
        obj.style.top  = style.top;
        ani[0].cancel()
    }

    obj.style.position = "relative"
    const ret = obj.animate([
        {
            left: obj.style.left || "0px",
            top: obj.style.top || "0px",
        },
        {
            left: "0px",
            top: y || "0px",
        }
    ], speed);

    ret.finished.then(() => {
        obj.style.left = "0px";
        obj.style.top = y; 
    }).catch(() => console.log("error"));

    return ret;
}

function drag(elem, container, targets) { 
    elem.style.userSelect = "none";
    const anim_speed = 500;
    var empty; 
    var off_x; 
    var off_y;

    elem.addEventListener("mousedown", start);
    elem.addEventListener("touchstart", start);

    function start(e) {
        e.preventDefault()
        elem.style.position = "relative";
        empty = elem.getBoundingClientRect()

        if (e.type == "mousedown") {
            off_x = e.clientX;
            off_y = e.clientY;
            document.addEventListener("mousemove", move);
            document.addEventListener("mouseup", stop);
            return 
        } 

        const evt = e.touches[0];
        off_x = evt.clientX; 
        off_y = evt.clientY;
        document.addEventListener("touchmove", move);
        document.addEventListener("touchend", stop);
    }

    function move(e) {
        const evt = (e.type != "touchmove")? e: e.touches[0]; 
        elem.style.left = (evt.clientX - off_x) + "px";
        elem.style.top = (evt.clientY - off_y) + "px";

        const elem_r = elem.getBoundingClientRect();
        if (collide(elem_r, empty)) {
            return
        }

        targets.forEach((target) => {
            if (target == elem) return
            const rect = target.getBoundingClientRect();
            if (!collide(elem_r, rect)) return

            if (empty.y > rect.y) {
                console.log("go up")
                const y = (target.style.top == "-60px")? "0px": "60px";
                transition(target, y, anim_speed)
                empty = rect;
                return;
            }
            console.log("go up")
            const y = (target.style.top == "60px")? "0px": "-60px";
            transition(target, y, anim_speed)
            empty = rect;
        })

    }

    function stop(e) {
        const move_ev = (e.type == 'mouseup')? "mousemove": "touchmove";
        document.removeEventListener(move_ev, move);
        document.removeEventListener(e.type, stop);

        const elem_r = elem.getBoundingClientRect();
        if (!collide(empty, elem_r)) {
            targets.forEach((target) => 
                transition(target, "0px", anim_speed));
            return
        }

        empty.y -= empty.height;
        const elem_id = elem.id;
        targets.forEach((target) => {
            const rect = target.getBoundingClientRect();
            if (collide(empty, rect) && target != elem) {
                console.log("times")
                elem.id = target.id;
                if (elem_id < target.id)  {
                    console.log("down", "priority", target.id)
                    fetch("/change_priority?task=" + 
                        elem.querySelector('button').value +
                        "&priority=" + target.id + "&direction=down" )
                        .then(res => res.text())
                        .then(bod => console.log(bod));

                    for (let i = elem.nextElementSibling; 
                        i != target.nextElementSibling; 
                        i = i.nextElementSibling) {
                        i.id--;
                    }
                    container.insertBefore(elem, target.nextElementSibling)
                }
                else  {
                    const next = (target.nextElementSibling == null)? 
                                        target: target.nextElementSibling;
                    fetch("/change_priority?task=" + 
                        elem.querySelector('button').value +
                        "&priority=" + next.id + "&direction=up") 
                        .then(res => res.text())
                        .then(bod => console.log(bod));
                    for (let i = target; i != elem;  i = i.nextElementSibling) {
                        i.id++;
                    }
                    container.insertBefore(elem, next) 
                }
            }
            target.style.top = ""
            target.style.position = ""
        })
    }
}

function collide(a, b) {
    return (b.x < a.x + a.width &&
            b.x + b.width > a.x) && 
           (b.y < a.y + a.height && 
            b.y + b.height > a.y) 
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
