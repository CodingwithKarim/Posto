const logoutLink = document.getElementById('logout-link');

if (logoutLink) {
    logoutLink.addEventListener('click', function(e) {
        // Prevent default behavior of anchor element
        e.preventDefault();  

        // Submit the hidden form
        document.getElementById('logout-form').submit();
    });
}