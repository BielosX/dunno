package org.dunno.model;

import java.time.LocalDate;
import java.util.List;
import java.util.UUID;

public record Books(List<Book> books, String requestId) {
  public record Book(UUID id, String title, String authors, LocalDate releaseDate) {}
}
