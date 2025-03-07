const likesCount = parseInt(document.getElementById('likeCount')?.getAttribute('data-count'), 10) || 0;

const setupLikesModalHover = () => {
  const likeButton = document.getElementById('likeButton');
  const likesModalEl = document.getElementById('likesModal');

  if (!likeButton || !likesModalEl) return;

  const likesModal = new bootstrap.Modal(likesModalEl);
  let hoverTimeout;

  if (likesCount > 0) {
    likeButton.addEventListener('mouseenter', () => {
      hoverTimeout = setTimeout(() => {
        likesModal.show();
      }, 1000);
    });
    likeButton.addEventListener('mouseleave', () => {
      clearTimeout(hoverTimeout);
    });
  }
};

const setupLikeToggle = () => {
  const likeButton = document.getElementById('likeButton');
  const likeIcon = document.getElementById('likeIcon');
  const likeCountSpan = document.getElementById('likeCount');

  if (!likeButton || !likeIcon || !likeCountSpan) return;

  let currentCount = parseInt(likeCountSpan.getAttribute('data-count'), 10) || 0;

  likeButton.addEventListener('click', async (e) => {
    console.log('Like button clicked');
    e.preventDefault();
    try {
      console.log('Like button clicked');
      const postID = likeButton.getAttribute('data-post-id');
      console.log(postID);
      const response = await fetch(`/blogpost/${postID}/like`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      });

      const data = await response.json();

      console.log(data);

      if (!response.ok) {
        console.error(data.error || 'Error toggling like');
        return;
      }

      if (data.liked) {
        currentCount++;
        likeIcon.className = 'fas fa-heart me-2';
      } else {
        currentCount--;
        likeIcon.className = 'far fa-heart me-2';
      }

      // Update both the data attribute and the visible text.
      likeCountSpan.setAttribute('data-count', currentCount);
      likeCountSpan.textContent = `${currentCount} ${currentCount === 1 ? "Like" : "Likes"}`;
    } catch (error) {
      console.error('Error:', error);
    }
  });
};

const init = () => {
    setupLikesModalHover();
    setupLikeToggle();
  };

init();
