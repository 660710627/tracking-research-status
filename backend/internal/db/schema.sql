CREATE TABLE IF NOT EXISTS researches (
    id INTEGER PRIMARY KEY AUTOINCREMENT CHECK (id > 0),
    title TEXT NOT NULL CHECK (research_text_valid(title, 1)),
    is_subsidized INTEGER NOT NULL CHECK (is_subsidized IN (0, 1)),
    project_type TEXT NOT NULL CHECK (project_type IN ('RESEARCH', 'ACADEMIC_SERVICE')),
    research_kind TEXT NOT NULL CHECK (research_kind IN ('BUDGET', 'CONTINUATION')),
    continuation_of_id INTEGER REFERENCES researches(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    responsible_project_unit TEXT NOT NULL CHECK (research_text_valid(responsible_project_unit, 0)),
    responsible_budget_unit TEXT NOT NULL CHECK (research_text_valid(responsible_budget_unit, 0)),
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    budget_amount REAL NOT NULL CHECK (research_decimal_valid(budget_amount)),
    thai_abstract TEXT NOT NULL CHECK (research_text_valid(thai_abstract, 0)),
    english_abstract TEXT NOT NULL CHECK (research_text_valid(english_abstract, 0)),
    objectives TEXT NOT NULL CHECK (research_text_valid(objectives, 0)),
    keywords TEXT NOT NULL CHECK (research_text_valid(keywords, 0)),
    status TEXT NOT NULL DEFAULT 'กำลังดำเนินการ' CHECK (status IN (
        'กำลังดำเนินการ', 'กำลังดำเนินการ(ขยายเวลาครั้งที่ 1)',
        'กำลังดำเนินการ(ขยายเวลาครั้งที่ 2)', 'กำลังดำเนินการ(ขยายเวลามากกว่า 2 ครั้ง)',
        'โครงการเสร็จสิ้น', 'ยุติโครงการ')),
    process TEXT NOT NULL DEFAULT 'สัญญาโครงการ' CHECK (process IN (
        'สัญญาโครงการ', 'บันทึกข้อตกลง', 'เปิดบัญชีธนาคาร', 'การเบิกจ่ายเงิน',
        'การจัดสรรค่าธรรมเนียม', 'การติดตามส่งรายงาน', 'รายงานสรุปการใช้เงิน', 'การปิดบัญชีธนาคาร')),
    CHECK ((research_kind = 'BUDGET' AND continuation_of_id IS NULL)
        OR (research_kind = 'CONTINUATION' AND continuation_of_id IS NOT NULL AND continuation_of_id > 0)),
    CHECK (research_dates_valid(start_date, end_date))
);

CREATE INDEX IF NOT EXISTS researches_title_key ON researches(research_trim(title));
CREATE INDEX IF NOT EXISTS researches_parent ON researches(continuation_of_id);

CREATE TRIGGER IF NOT EXISTS researches_root_title_insert
BEFORE INSERT ON researches
WHEN NEW.research_kind = 'BUDGET' AND EXISTS (
    SELECT 1 FROM researches WHERE research_trim(title) = research_trim(NEW.title)
)
BEGIN
    SELECT RAISE(ABORT, 'research_title_already_exists');
END;

CREATE TRIGGER IF NOT EXISTS researches_identity_immutable
BEFORE UPDATE OF id, continuation_of_id, research_kind ON researches
WHEN NEW.id IS NOT OLD.id OR NEW.continuation_of_id IS NOT OLD.continuation_of_id
    OR NEW.research_kind IS NOT OLD.research_kind
BEGIN
    SELECT RAISE(ABORT, 'research_constraint_identity_immutable');
END;

CREATE TABLE IF NOT EXISTS research_members (
    research_id INTEGER NOT NULL REFERENCES researches(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    full_name TEXT NOT NULL CHECK (research_text_valid(full_name, 0)),
    email TEXT NOT NULL CHECK (research_email_valid(email)),
    affiliation TEXT NOT NULL CHECK (research_text_valid(affiliation, 0)),
    contribution_percent REAL NOT NULL CHECK (
        research_decimal_valid(contribution_percent) AND contribution_percent <= 100),
    role TEXT NOT NULL CHECK (role IN ('LEAD', 'CO_RESEARCHER')),
    PRIMARY KEY (research_id, role)
);

CREATE UNIQUE INDEX IF NOT EXISTS research_members_email_unique
ON research_members(research_id, research_email_key(email));

CREATE TABLE IF NOT EXISTS research_contracts (
    research_id INTEGER PRIMARY KEY REFERENCES researches(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    funding_type TEXT NOT NULL CHECK (funding_type IN ('INTERNAL', 'EXTERNAL')),
    funding_source_name TEXT NOT NULL CHECK (research_text_valid(funding_source_name, 0)),
    contract_number TEXT NOT NULL CHECK (research_text_valid(contract_number, 0)),
    contract_number_key TEXT NOT NULL UNIQUE CHECK (length(contract_number_key) > 0),
    storage_path TEXT NOT NULL CHECK (length(research_trim(storage_path)) > 0),
    original_filename TEXT NOT NULL CHECK (research_text_valid(original_filename, 0)),
    content_type TEXT NOT NULL CHECK (content_type = 'application/pdf'),
    size_bytes INTEGER NOT NULL CHECK (typeof(size_bytes) = 'integer' AND size_bytes BETWEEN 1 AND 20971520)
);

-- Create inserts the contract last, sealing the complete personnel aggregate.
-- Unique roles plus count=2 require one LEAD and one CO_RESEARCHER. Each
-- individual value has already passed the decimal CHECK before cents are summed.
CREATE TRIGGER IF NOT EXISTS research_contract_members_complete
BEFORE INSERT ON research_contracts
WHEN (SELECT count(*) FROM research_members WHERE research_id = NEW.research_id) <> 2
    OR (SELECT sum(CAST(round(contribution_percent * 100) AS INTEGER))
        FROM research_members WHERE research_id = NEW.research_id) <> 10000
BEGIN
    SELECT RAISE(ABORT, 'research_constraint_members_incomplete');
END;
