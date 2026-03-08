export async function toggleFavorite(form: HTMLFormElement) {
  form.requestSubmit();
}

export async function deleteArticle(form: HTMLFormElement) {
  if (!window.confirm('Are you sure you want to delete this article?')) {
    return;
  }
  form.requestSubmit();
}

export async function sendArticle(form: HTMLFormElement) {
  form.requestSubmit();
}
