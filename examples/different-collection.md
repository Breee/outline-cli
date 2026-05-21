<!-- Collection: Welcome -->
<!-- Title: Runbook: Incident Response -->

# Runbook: Incident Response

## Severity Levels

| Level | Response Time | Escalation |
|-------|--------------|------------|
| P1    | 15 min       | Immediate  |
| P2    | 1 hour       | Team lead  |
| P3    | 4 hours      | Next standup |
| P4    | Best effort  | Backlog    |

## Steps

1. Acknowledge the alert
2. Assess impact and assign severity
3. Open incident channel
4. Investigate root cause
5. Mitigate
6. Write post-mortem

## Useful Commands

```bash
kubectl get pods -A | grep -v Running
kubectl logs -f deployment/api --since=5m
```
