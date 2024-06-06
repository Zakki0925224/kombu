use crate::{
    detection_rule::DetectionRule, sandbox::SandboxResult, syscall::X86_64Syscalls, uptime::Uptime,
};
use anyhow::{anyhow, Result};
use serde::Deserialize;
use std::fs;

pub type SyscallEvents = Vec<SyscallEvent>;

#[derive(Deserialize, Debug)]
pub struct SyscallEvent {
    timestamp: u64,
    nr: u32,
    pid: u32,
    ppid: u32,
    comm: [u8; 16],
}

impl SyscallEvent {
    fn comm_to_string(&self) -> String {
        String::from_utf8_lossy(&self.comm)
            .to_string()
            .replace('\0', "")
    }

    fn nr_to_syscall(&self) -> Option<X86_64Syscalls> {
        if let Ok(s) = X86_64Syscalls::try_from(self.nr) {
            Some(s)
        } else {
            None
        }
    }

    fn timestamp_to_uptime(&self) -> Uptime {
        Uptime::from(self.timestamp)
    }
}

pub struct AnalyzeResult {
    pub violated_detection_rule: Option<DetectionRule>,
    pub message: String,
}

pub struct Analyzer {
    sandbox_result: SandboxResult,
    detection_rules: Vec<DetectionRule>,
}

impl Analyzer {
    pub fn new(sandbox_result: SandboxResult, detection_rules_dir_path: String) -> Result<Self> {
        let mut detection_rules = Vec::new();

        // parse detection rules
        let entries = fs::read_dir(detection_rules_dir_path)?;
        for e in entries {
            let f = fs::read_to_string(e?.path())?;
            let rule = toml::from_str(&f)?;
            detection_rules.push(rule);
        }

        if detection_rules.len() == 0 {
            return Err(anyhow!("Not a single detection rule was found"));
        }

        Ok(Self {
            sandbox_result,
            detection_rules,
        })
    }

    pub fn analyze(&self) -> Result<AnalyzeResult> {
        let filtered = self
            .sandbox_result
            .syscall_events
            .iter()
            .filter(|e| e.comm_to_string() == "target")
            .collect::<Vec<_>>();

        for rule in &self.detection_rules {
            if let Some(syscall) = &rule.syscall {
                // check blacklist
                for num in &syscall.blacklist_numbers {
                    for e in &filtered {
                        if e.nr as usize == *num {
                            return Ok(AnalyzeResult {
                                violated_detection_rule: Some(rule.clone()),
                                message: format!("Blacklisted syscall detected: {:?}", num),
                            });
                        }
                    }
                }

                // check frequent
                for f in &syscall.frequent {
                    let threshold = f.threshold_count;
                    if threshold
                        <= filtered
                            .iter()
                            .filter(|e| e.nr as usize == f.number)
                            .count()
                    {
                        return Ok(AnalyzeResult {
                            violated_detection_rule: Some(rule.clone()),
                            message: format!("Frequent syscall detected: {:?}", f.number),
                        });
                    }
                }

                // check consecutive
                for c in &syscall.consecutive {
                    let threshold = c.threshold_count;
                    let mut count = 0;
                    for e in &filtered {
                        if e.nr as usize == c.number {
                            count += 1;
                        } else {
                            count = 0;
                        }
                    }

                    if count >= threshold {
                        return Ok(AnalyzeResult {
                            violated_detection_rule: Some(rule.clone()),
                            message: format!("Consecutive syscall detected: {:?}", c.number),
                        });
                    }
                }
            }

            if let Some(timestamp) = &rule.timestamp {
                // check timetravel
                if timestamp.check_timetravel {
                    // to past
                    let mut prev_timestamp = None;
                    for e in &filtered {
                        if let Some(prev) = prev_timestamp {
                            if prev > e.timestamp {
                                return Ok(AnalyzeResult {
                                    violated_detection_rule: Some(rule.clone()),
                                    message: format!(
                                        "Timetravel detected: prev: {} > next: {}",
                                        prev, e.timestamp
                                    ),
                                });
                            }
                        }
                        prev_timestamp = Some(e.timestamp);
                    }

                    // to future (1s)
                    let mut prev_timestamp = None;
                    for e in &filtered {
                        if let Some(prev) = prev_timestamp {
                            if e.timestamp - prev >= 1_000_000_000 {
                                return Ok(AnalyzeResult {
                                    violated_detection_rule: Some(rule.clone()),
                                    message: format!(
                                        "Timetravel detected: prev: {} > next: {}",
                                        prev, e.timestamp
                                    ),
                                });
                            }
                        }
                        prev_timestamp = Some(e.timestamp);
                    }
                }
            }
        }

        // for e in &filtered {
        //     println!(
        //         "uptime: {}, syscall: {:?}, pid: {}, ppid: {}, comm: {:?}",
        //         e.timestamp_to_uptime(),
        //         e.nr_to_syscall(),
        //         e.pid,
        //         e.ppid,
        //         e.comm_to_string()
        //     );
        // }
        // println!("events count: {}", self.sandbox_result.syscall_events.len());

        Ok(AnalyzeResult {
            violated_detection_rule: None,
            message: "".to_string(),
        })
    }
}
