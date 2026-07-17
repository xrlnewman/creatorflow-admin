-- CreatorFlow synthetic operational data only. Never load real medical records here.
CREATE TABLE IF NOT EXISTS departments (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS doctors (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  department VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  today_count INT NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS patients (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  phone VARCHAR(32) NOT NULL,
  last_visit VARCHAR(32) NOT NULL,
  created_at VARCHAR(64) NOT NULL
);
CREATE TABLE IF NOT EXISTS appointments (
  id VARCHAR(64) PRIMARY KEY,
  patient_id VARCHAR(64) NOT NULL,
  patient_name VARCHAR(64) NOT NULL,
  department VARCHAR(64) NOT NULL,
  doctor VARCHAR(64) NOT NULL,
  scheduled_at VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  created_at VARCHAR(64) NOT NULL,
  updated_at VARCHAR(64) NOT NULL,
  INDEX idx_appointments_status_time (status, scheduled_at)
);
CREATE TABLE IF NOT EXISTS appointment_events (
  id VARCHAR(64) PRIMARY KEY,
  appointment_id VARCHAR(64) NOT NULL,
  from_status VARCHAR(32) NOT NULL,
  to_status VARCHAR(32) NOT NULL,
  actor VARCHAR(64) NOT NULL,
  created_at VARCHAR(64) NOT NULL,
  INDEX idx_appointment_events_appointment (appointment_id, created_at)
);
CREATE TABLE IF NOT EXISTS followups (
  id VARCHAR(64) PRIMARY KEY,
  patient_id VARCHAR(64) NOT NULL,
  patient_name VARCHAR(64) NOT NULL,
  summary VARCHAR(255) NOT NULL,
  due_at VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  created_at VARCHAR(64) NOT NULL,
  updated_at VARCHAR(64) NOT NULL,
  INDEX idx_followups_status_due (status, due_at)
);
CREATE TABLE IF NOT EXISTS content_items (
  id VARCHAR(64) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  channel VARCHAR(64) NOT NULL,
  owner VARCHAR(64) NOT NULL,
  planned_at VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  created_at VARCHAR(64) NOT NULL,
  updated_at VARCHAR(64) NOT NULL,
  INDEX idx_content_items_status_plan (status, planned_at),
  INDEX idx_content_items_owner_plan (owner, planned_at)
);
CREATE TABLE IF NOT EXISTS content_scripts (
  id VARCHAR(64) PRIMARY KEY,
  content_item_id VARCHAR(64) NOT NULL UNIQUE,
  body TEXT NOT NULL,
  updated_at VARCHAR(64) NOT NULL,
  CONSTRAINT fk_content_scripts_item FOREIGN KEY (content_item_id) REFERENCES content_items(id)
);
CREATE TABLE IF NOT EXISTS content_publish_records (
  id VARCHAR(64) PRIMARY KEY,
  content_item_id VARCHAR(64) NOT NULL UNIQUE,
  published_at VARCHAR(64) NOT NULL,
  actor VARCHAR(64) NOT NULL,
  created_at VARCHAR(64) NOT NULL,
  INDEX idx_content_publish_time (published_at),
  CONSTRAINT fk_content_publish_item FOREIGN KEY (content_item_id) REFERENCES content_items(id)
);
CREATE TABLE IF NOT EXISTS content_metrics (
  id VARCHAR(64) PRIMARY KEY,
  content_item_id VARCHAR(64) NOT NULL UNIQUE,
  views BIGINT NOT NULL DEFAULT 0,
  likes BIGINT NOT NULL DEFAULT 0,
  comments BIGINT NOT NULL DEFAULT 0,
  shares BIGINT NOT NULL DEFAULT 0,
  recorded_at VARCHAR(64) NOT NULL,
  CONSTRAINT chk_content_metrics_non_negative CHECK (views >= 0 AND likes >= 0 AND comments >= 0 AND shares >= 0),
  CONSTRAINT fk_content_metrics_item FOREIGN KEY (content_item_id) REFERENCES content_items(id)
);
CREATE TABLE IF NOT EXISTS content_events (
  id VARCHAR(64) PRIMARY KEY,
  content_item_id VARCHAR(64) NOT NULL,
  from_status VARCHAR(32) NOT NULL DEFAULT '',
  to_status VARCHAR(32) NOT NULL,
  action VARCHAR(64) NOT NULL,
  actor VARCHAR(64) NOT NULL,
  created_at VARCHAR(64) NOT NULL,
  INDEX idx_content_events_item_time (content_item_id, created_at),
  CONSTRAINT fk_content_events_item FOREIGN KEY (content_item_id) REFERENCES content_items(id)
);

INSERT IGNORE INTO departments (id,name) VALUES
 ('dep-video','短视频'),('dep-article','图文专栏'),('dep-live','直播栏目'),('dep-brand','品牌合作');
INSERT IGNORE INTO doctors (id,name,department,status,today_count) VALUES
 ('doc-01','林编辑','短视频','排期中',18),('doc-02','沈编辑','图文专栏','排期中',16),
 ('doc-03','赵编辑','直播栏目','制作中',12),('doc-04','周编辑','品牌合作','休息中',10),
 ('doc-05','陈编辑','短视频','排期中',14),('doc-06','王编辑','图文专栏','排期中',16);
INSERT IGNORE INTO patients (id,name,phone,last_visit,created_at) VALUES
 ('PT-001','演示创作者01','13800000001','2026-07-15','2026-07-01'),('PT-002','演示创作者02','13800000002','2026-07-15','2026-07-01'),
 ('PT-003','演示创作者03','13800000003','2026-07-14','2026-07-01'),('PT-004','演示创作者04','13800000004','2026-07-14','2026-07-01'),
 ('PT-005','演示创作者05','13800000005','2026-07-13','2026-07-01'),('PT-006','演示创作者06','13800000006','2026-07-13','2026-07-01'),
 ('PT-007','演示创作者07','13800000007','2026-07-12','2026-07-01'),('PT-008','演示创作者08','13800000008','2026-07-12','2026-07-01'),
 ('PT-009','演示创作者09','13800000009','2026-07-11','2026-07-01'),('PT-010','演示创作者10','13800000010','2026-07-11','2026-07-01'),
 ('PT-011','演示创作者11','13800000011','2026-07-10','2026-07-01'),('PT-012','演示创作者12','13800000012','2026-07-10','2026-07-01'),
 ('PT-013','演示创作者13','13800000013','2026-07-09','2026-07-01'),('PT-014','演示创作者14','13800000014','2026-07-09','2026-07-01'),
 ('PT-015','演示创作者15','13800000015','2026-07-08','2026-07-01'),('PT-016','演示创作者16','13800000016','2026-07-08','2026-07-01'),
 ('PT-017','演示创作者17','13800000017','2026-07-07','2026-07-01'),('PT-018','演示创作者18','13800000018','2026-07-07','2026-07-01'),
 ('PT-019','演示创作者19','13800000019','2026-07-06','2026-07-01'),('PT-020','演示创作者20','13800000020','2026-07-06','2026-07-01'),
 ('PT-021','演示创作者21','13800000021','2026-07-05','2026-07-01'),('PT-022','演示创作者22','13800000022','2026-07-05','2026-07-01'),
 ('PT-023','演示创作者23','13800000023','2026-07-04','2026-07-01'),('PT-024','演示创作者24','13800000024','2026-07-04','2026-07-01'),
 ('PT-025','演示创作者25','13800000025','2026-07-03','2026-07-01'),('PT-026','演示创作者26','13800000026','2026-07-03','2026-07-01'),
 ('PT-027','演示创作者27','13800000027','2026-07-02','2026-07-01'),('PT-028','演示创作者28','13800000028','2026-07-02','2026-07-01'),
 ('PT-029','演示创作者29','13800000029','2026-07-01','2026-07-01'),('PT-030','演示创作者30','13800000030','2026-07-01','2026-07-01');
INSERT IGNORE INTO appointments (id,patient_id,patient_name,department,doctor,scheduled_at,status,created_at,updated_at) VALUES
 ('CR-0716-081','PT-001','演示创作者01','短视频','林编辑','2026-07-16T08:00:00+08:00','已发布','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-082','PT-002','演示创作者02','图文专栏','沈编辑','2026-07-16T09:00:00+08:00','制作中','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-083','PT-003','演示创作者03','直播栏目','赵编辑','2026-07-16T10:00:00+08:00','待制作','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-084','PT-004','演示创作者04','品牌合作','周编辑','2026-07-16T11:00:00+08:00','已排期','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-085','PT-005','演示创作者05','短视频','陈编辑','2026-07-16T12:00:00+08:00','待排期','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-086','PT-006','演示创作者06','图文专栏','王编辑','2026-07-16T13:00:00+08:00','已发布','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-087','PT-007','演示创作者07','直播栏目','赵编辑','2026-07-16T14:00:00+08:00','制作中','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-088','PT-008','演示创作者08','品牌合作','周编辑','2026-07-16T15:00:00+08:00','待制作','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-089','PT-009','演示创作者09','短视频','林编辑','2026-07-16T16:00:00+08:00','已排期','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-090','PT-010','演示创作者10','图文专栏','沈编辑','2026-07-16T17:00:00+08:00','待排期','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-091','PT-011','演示创作者11','直播栏目','赵编辑','2026-07-16T08:30:00+08:00','已发布','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z'),
 ('CR-0716-092','PT-012','演示创作者12','品牌合作','周编辑','2026-07-16T09:30:00+08:00','待制作','2026-07-16T00:00:00Z','2026-07-16T01:00:00Z');
INSERT IGNORE INTO followups (id,patient_id,patient_name,summary,due_at,status,created_at,updated_at) VALUES
 ('RV-0716-001','PT-001','演示创作者01','标题与封面复核','2026-07-17','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-002','PT-002','演示创作者02','素材版权检查','2026-07-17','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-003','PT-003','演示创作者03','发布数据复盘','2026-07-18','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-004','PT-004','演示创作者04','评论区互动复盘','2026-07-18','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-005','PT-005','演示创作者05','选题热度复盘','2026-07-19','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-006','PT-006','演示创作者06','品牌合作复盘','2026-07-19','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-007','PT-007','演示创作者07','直播数据复盘','2026-07-20','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-008','PT-008','演示创作者08','素材归档提醒','2026-07-20','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-009','PT-009','演示创作者09','评论区复盘记录','2026-07-21','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-010','PT-010','演示创作者10','内容质量抽检','2026-07-21','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-011','PT-011','演示创作者11','发布节奏复盘','2026-07-22','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('RV-0716-012','PT-012','演示创作者12','粉丝增长复盘','2026-07-22','待完成','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z');
INSERT IGNORE INTO content_items (id,title,channel,owner,planned_at,status,created_at,updated_at) VALUES
 ('CF-0718-001','城市夜行：下班后的十五分钟','短视频','林编辑','2026-07-18T09:00:00+08:00','待选题','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z'),
 ('CF-0718-002','一周好物：把桌面整理成工作流','图文专栏','沈编辑','2026-07-18T10:00:00+08:00','写作中','2026-07-16T00:00:00Z','2026-07-16T02:00:00Z'),
 ('CF-0718-003','品牌访谈：小店如何留住老客','直播栏目','赵编辑','2026-07-18T14:00:00+08:00','制作中','2026-07-16T00:00:00Z','2026-07-16T03:00:00Z'),
 ('CF-0718-004','夏日直播：创作者增长公开课','品牌合作','周编辑','2026-07-18T16:00:00+08:00','待审核','2026-07-16T00:00:00Z','2026-07-16T04:00:00Z'),
 ('CF-0718-005','通勤装备：轻量化出行清单','短视频','林编辑','2026-07-18T18:00:00+08:00','已发布','2026-07-16T00:00:00Z','2026-07-16T05:00:00Z'),
 ('CF-0718-006','一张图读懂内容复盘','图文专栏','沈编辑','2026-07-19T09:00:00+08:00','已复盘','2026-07-16T00:00:00Z','2026-07-16T06:00:00Z');
INSERT IGNORE INTO content_scripts (id,content_item_id,body,updated_at) VALUES
 ('CF-0718-002-SCRIPT','CF-0718-002','开场钩子、三段主体和结尾行动号召。','2026-07-16T02:00:00Z'),
 ('CF-0718-003-SCRIPT','CF-0718-003','开场钩子、三段主体和结尾行动号召。','2026-07-16T03:00:00Z'),
 ('CF-0718-004-SCRIPT','CF-0718-004','开场钩子、三段主体和结尾行动号召。','2026-07-16T04:00:00Z'),
 ('CF-0718-005-SCRIPT','CF-0718-005','开场钩子、三段主体和结尾行动号召。','2026-07-16T05:00:00Z'),
 ('CF-0718-006-SCRIPT','CF-0718-006','开场钩子、三段主体和结尾行动号召。','2026-07-16T06:00:00Z');
INSERT IGNORE INTO content_publish_records (id,content_item_id,published_at,actor,created_at) VALUES
 ('CF-0718-005-PUB','CF-0718-005','2026-07-18T18:00:00+08:00','主编','2026-07-16T05:00:00Z'),
 ('CF-0718-006-PUB','CF-0718-006','2026-07-18T18:00:00+08:00','主编','2026-07-16T06:00:00Z');
INSERT IGNORE INTO content_metrics (id,content_item_id,views,likes,comments,shares,recorded_at) VALUES
 ('CF-0718-006-METRIC','CF-0718-006',12480,892,67,141,'2026-07-16T06:30:00Z');
INSERT IGNORE INTO content_events (id,content_item_id,from_status,to_status,action,actor,created_at) VALUES
 ('CF-0718-001-EV-1','CF-0718-001','','待选题','create','林编辑','2026-07-16T00:00:00Z'),
 ('CF-0718-002-EV-1','CF-0718-002','','待选题','create','沈编辑','2026-07-16T00:00:00Z'),
 ('CF-0718-002-EV-2','CF-0718-002','待选题','写作中','write_script','沈编辑','2026-07-16T02:00:00Z'),
 ('CF-0718-003-EV-1','CF-0718-003','','待选题','create','赵编辑','2026-07-16T00:00:00Z'),
 ('CF-0718-003-EV-2','CF-0718-003','待选题','写作中','write_script','赵编辑','2026-07-16T01:00:00Z'),
 ('CF-0718-003-EV-3','CF-0718-003','写作中','制作中','start_production','赵编辑','2026-07-16T03:00:00Z'),
 ('CF-0718-004-EV-1','CF-0718-004','制作中','待审核','submit_review','周编辑','2026-07-16T04:00:00Z'),
 ('CF-0718-005-EV-1','CF-0718-005','待审核','已发布','publish','主编','2026-07-16T05:00:00Z'),
 ('CF-0718-006-EV-1','CF-0718-006','已发布','已复盘','record_metrics','运营人员','2026-07-16T06:30:00Z');
