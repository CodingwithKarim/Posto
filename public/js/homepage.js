// Get the current username and ownership status from data attributes
const username = document.body.dataset.username;
const isOwner = document.body.dataset.owner === "true"; // Ensure it's treated as a boolean

let currentPage = 1; // Default to the first page
const postsContainer = document.getElementById("posts-wrapper");
const showMoreButton = document.getElementById("show-more");

// Add a click event listener to the "Older Posts" button
if (showMoreButton){
    showMoreButton.addEventListener("click", function () {
        currentPage += 1; // Increment the page number
        loadPosts(currentPage);
    });
}

// Function to load posts for the given user and page
function loadPosts(page) {
    fetch(`/profile/${username}/?page=${page}`, {
        method: 'GET',
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
    .then(response => response.json())
    .then(posts => {
        if (posts && posts.length > 0) {
            appendPosts(posts);
            // Hide the "Older Posts" button if fewer than expected posts are returned
            if (posts.length < 3) showMoreButton.style.display = "none";
        } else {
            // No more posts, hide the button
            showMoreButton.style.display = "none";
        }
    })
    .catch(error => {
        console.error("Error fetching posts:", error);
    });
}

// Function to append posts to the DOM
function appendPosts(posts) {
    posts.forEach(post => {
        const postElement = document.createElement("div");
        postElement.classList.add("post-preview");
        postElement.dataset.postId = post.ID;
        postElement.innerHTML = `
            <a href="/blogpost/${post.ID}">
                <h2 class="post-title post-title-page">${post.Title}</h2>
                <h3 class="post-subtitle post-subtitle-page">${post.Content}</h3>
            </a>
            <p class="post-meta">Posted by ${username} on ${post.CreatedAt}</p>
            ${isOwner ? `
            <div class="post-actions">
                <a href="/edit/${post.ID}" class="edit-link">
                    <i class="fas fa-edit"></i> Edit
                </a>
                <form action="/delete/${post.ID}" method="POST" class="delete-form" onsubmit="return confirm('Are you sure you want to delete this post?');">
                    <button type="submit" class="delete-button">
                        <i class="fas fa-trash-alt"></i> Delete
                    </button>
                </form>
            </div>` : ''}
        `;

        postsContainer.appendChild(postElement);

        // Add a divider after each post
        const divider = document.createElement("hr");
        divider.classList.add("my-4");
        postsContainer.appendChild(divider);
    });
}