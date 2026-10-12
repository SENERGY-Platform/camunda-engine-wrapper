# Switching a deployment off and on

## Scope

Holds for process deployments managed by this wrapper (cloud), from `v1.1.0`. Fog deployments on a gateway are not covered; they go through process-sync and never reach the wrapper.

Not about deleting a deployment: switching off keeps the deployment and its process model, deleting removes both.

## Behaviour

`PUT /v2/deployments/{id}/disabled` with `{"disabled": true|false}` switches the deployment of the calling user (`401` for another tenant or an unknown id, `400` without a boolean `disabled`).

- **Off:** the intent is stored, the Camunda definitions of the deployment are suspended, then its running instances are removed (process-io variables of the instances are deleted with them). A suspended definition blocks manual, timer and event starts; a message correlation with `all: true` answers `204` and starts nothing.
- **On:** the definitions are activated, then the intent is removed. If activation fails the intent stays, so a retry completes it.
- `GET /v2/deployments` reports `disabled` per deployment, read from the definition's `suspended` flag, so it shows the state Camunda enforces.
- All start endpoints (`/v2/deployments/{id}/start`, `/deployment/{id}/start`, and the definition-based starts) answer `409` for a suspended definition.

## Persistence across redeploys

A redeploy (`PUT /process-deployments`) deletes the old Camunda deployment and creates a new definition, which Camunda creates active. The intent is therefore kept in the table `DisabledDeployment` (keyed by the virtual deployment id) in the wrapper database:

- `Deploy` applies the intent right after the Camunda deploy, before incident handling and the vid relation, to keep the window in which the new definition is active short. It checks again after the vid relation is saved, for a disable that ran concurrently.
- If applying the intent fails, the new Camunda deployment is removed and the deploy answers `500`.
- The redeploy cleanup keeps the intent; a real `DELETE /process-deployments/{userid}/{deplid}` removes it.

## Known limitations

- Between the Camunda deploy and the suspension there are two REST round trips. An instance started in that window is removed by the same step, but a worker may fetch its first external task before that.
- Enable, disable and redeploy of the same id running at the same time can leave the intent and the definition state apart. Repeating the switch call fixes it; there is no per-deployment lock.
- A suspended definition keeps its message start subscription, so another deployment of the same tenant with the same message name still fails to deploy and falls back to the placeholder process.
