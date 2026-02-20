<section>
	<hgroup>
		<h1>Add Article</h1>
		<p>Save a new article to your reading list</p>
	</hgroup>

	<form method="POST" action="?/new">
		<fieldset>
			<label>
				URL
				<input
					type="url"
					name="url"
					required
					placeholder="https://example.com/article"
					autocomplete="url"
				/>
				<small>Enter the full URL of the article you want to save</small>
			</label>
		</fieldset>
		<button type="submit">Add</button>
	</form>
</section>
