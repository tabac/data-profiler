CREATE TABLE IF NOT EXISTS `datasets` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`path` VARCHAR(200),
		`name` VARCHAR(200),
		`description` VARCHAR(2000)
);


CREATE TABLE IF NOT EXISTS `matrices` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`path` VARCHAR(500),
		`filename` VARCHAR(500),
		`configuration` VARCHAR(2000),
		`datasetid` INTEGER,
		FOREIGN KEY(datasetid) REFERENCES datasets(id)
);

CREATE TABLE IF NOT EXISTS `estimators` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`path` VARCHAR(500),
		`filename` VARCHAR(500),
		`configuration` VARCHAR(2000),
		`datasetid` INTEGER,
		`matrixid` INTEGER,
		FOREIGN KEY(datasetid) REFERENCES datasets(id),
		FOREIGN KEY(matrixid) REFERENCES matrices(id)
);

CREATE TABLE IF NOT EXISTS `coordinates` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`path` VARCHAR(500),
		`filename` VARCHAR(500),
		`k` VARCHAR(500),
		`gof` VARCHAR(2000),
		`matrixid` INTEGER,
		FOREIGN KEY(matrixid) REFERENCES matrices(id)
);

CREATE TABLE IF NOT EXISTS `operators` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`name` VARCHAR(500),
		`description` VARCHAR(500),
		`path` VARCHAR(500),
		`datasetid` INTEGER,
		FOREIGN KEY(datasetid) REFERENCES datasets(id)
);

CREATE TABLE IF NOT EXISTS `scores` (
		`id` INTEGER PRIMARY KEY AUTOINCREMENT,
		`path` VARCHAR(500),
		`filename` VARCHAR(500),
		`operatorid` INTEGER,
		FOREIGN KEY(operatorid) REFERENCES operators(id)
);

