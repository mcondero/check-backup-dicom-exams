package model

type Config struct {
	CSVPath    string `yaml:"csv_path"`
	BaseDir    string `yaml:"base_dir"`
	OutputCSV  string `yaml:"output_csv"`
	SummaryTXT string `yaml:"summary_txt"`
}

type ExamRecord struct {
	FoundPath string
	DCMCount  int
	PDFCount  int
	ID        string
	ExamID    string
	PatientID string
	StudyUID  string
}
