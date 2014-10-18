/*
 This file was generated by Dashcode.
 You may edit this file to customize your widget or web page
 according to the license.txt file included in the project.
 */


//
// Function: updateGEISPhoto()
// Called by load and show to update photo
//
//
function updateGEISPhoto()
{
    var image = new Image();
    image.src = '/Users/dnewell/dev/geis-wallpaper/wallpaper.png?' + new Date().getTime();
    document.getElementById("geisWallpaper").src = image.src;
}

//
// Function: load()
// Called by HTML body element's onload event when the widget is ready to start
//
function load()
{
    dashcode.setupParts();

    setTimeout(function () { updateGEISPhoto(); }, 1000);
}

//
// Function: hide()
// Called when the widget has been hidden
//
function hide()
{
}

//
// Function: show()
// Called when the widget has been shown
//
function show()
{
    updateGEISPhoto();
}

//
// Function: showBack(event)
// Called when the info button is clicked to show the back of the widget
//
// event: onClick event from the info button
//
function showBack(event)
{
    var front = document.getElementById("front");
    var back = document.getElementById("back");

    if (window.widget) {
        widget.prepareForTransition("ToBack");
    }

    front.style.display="none";
    back.style.display="block";

    if (window.widget) {
        setTimeout('widget.performTransition();', 0);
    }
}

//
// Function: showFront(event)
// Called when the done button is clicked from the back of the widget
//
// event: onClick event from the done button
//
function showFront(event)
{
    var front = document.getElementById("front");
    var back = document.getElementById("back");

    if (window.widget) {
        widget.prepareForTransition("ToFront");
    }

    front.style.display="block";
    back.style.display="none";

    if (window.widget) {
        setTimeout('widget.performTransition();', 0);
    }
}

// Initialize the Dashboard event handlers
if (window.widget) {
    widget.onhide = hide;
    widget.onshow = show;
}
