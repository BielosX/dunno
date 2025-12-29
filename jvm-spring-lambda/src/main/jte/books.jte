@param org.dunno.model.Books books

<html lang="en">
    <body>
        <h2>Add new Book</h2>
        <form method="post" action="">
            <label>
                Title
                <input type="text" id="title" name="title">
            </label>
            <label>
                Authors
                <input type="text" id="authors" name="authors">
            </label>
            <label>
                Release Date
                <input type="date" id="releaseDate" name="releaseDate">
            </label>
            <button type="submit">Save</button>
        </form>
        <h2>Books</h2>
        <table>
            <tr>
                <th>Id</th>
                <th>Title</th>
                <th>Authors</th>
                <th>Release Date</th>
            </tr>
            @for(var book: books.books())
                <tr>
                    <th>${book.id().toString()}</th>
                    <th>${book.title()}</th>
                    <th>${book.authors()}</th>
                    <th>${book.releaseDate().toString()}</th>
                </tr>
            @endfor
        </table>
        <footer>
            <p>RequestId: ${books.requestId()}</p>
        </footer>
    </body>
</html>