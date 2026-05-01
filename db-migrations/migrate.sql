CREATE TABLE IF NOT EXISTS genre (
    id SMALLINT PRIMARY KEY,
    title VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS director (
    id SMALLINT PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS movie (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    year_released SMALLINT,
    rating NUMERIC(3,2),
    duration_mins SMALLINT,
    genre_id SMALLINT REFERENCES genre(id),
    director_id SMALLINT REFERENCES director(id)
);

CREATE INDEX IF NOT EXISTS movie_rating ON movie(rating);
CREATE INDEX IF NOT EXISTS movie_genre_id ON movie(genre_id);
CREATE INDEX IF NOT EXISTS movie_director_id ON movie(director_id);
CREATE INDEX IF NOT EXISTS movie_year_released ON movie(year_released);
CREATE INDEX IF NOT EXISTS movie_duration_mins ON movie(duration_mins);

INSERT INTO genre (id, title) VALUES
(1, 'Action'),
(2, 'Comedy'),
(3, 'Drama'),
(4, 'Horror'),
(5, 'SciFi'),
(6, 'Romance'),
(7, 'Thriller'),
(8, 'Animation'),
(9, 'Documentary'),
(10, 'Fantasy');

INSERT INTO director (id, first_name, last_name) VALUES
(1, 'Christopher', 'Nolan'),
(2, 'Quentin', 'Tarantino'),
(3, 'Martin', 'Scorsese'),
(4, 'Steven', 'Spielberg'),
(5, 'Ridley', 'Scott'),
(6, 'David', 'Fincher'),
(7, 'James', 'Cameron'),
(8, 'Denis', 'Villeneuve'),
(9, 'Greta', 'Gerwig'),
(10, 'Jordan', 'Peele');

INSERT INTO movie (title, year_released, rating, duration_mins, genre_id, director_id) VALUES
('Inception', 2010, 4.80, 148, 5, 1),
('The Dark Knight', 2008, 4.90, 152, 1, 1),
('Interstellar', 2014, 4.70, 169, 5, 1),
('Pulp Fiction', 1994, 4.85, 154, 7, 2),
('Inglourious Basterds', 2009, 4.60, 153, 2, 2),
('Goodfellas', 1990, 4.88, 146, 3, 3),
('The Wolf of Wall Street', 2013, 4.55, 180, 3, 3),
('Schindlers List', 1993, 4.92, 195, 3, 4),
('Jurassic Park', 1993, 4.50, 127, 5, 4),
('Gladiator', 2000, 4.62, 155, 1, 5),
('The Martian', 2015, 4.45, 144, 5, 5),
('Fight Club', 1999, 4.78, 139, 7, 6),
('Gone Girl', 2014, 4.52, 149, 7, 6),
('Titanic', 1997, 4.40, 194, 6, 7),
('Avatar', 2009, 4.20, 162, 5, 7),
('Arrival', 2016, 4.65, 116, 5, 8),
('Dune', 2021, 4.58, 155, 5, 8),
('Blade Runner 2049', 2017, 4.55, 164, 5, 8),
('Lady Bird', 2017, 4.30, 94, 3, 9),
('Barbie', 2023, 4.10, 114, 2, 9),
('Get Out', 2017, 4.48, 104, 4, 10),
('Us', 2019, 4.20, 116, 4, 10);