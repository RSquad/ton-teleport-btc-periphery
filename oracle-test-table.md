# Test cases

1. Deployment and basic checks

| Description                                                         | Status            | Comment                                                                 |
| :------------------------------------------------------------------ | :---------------- | :---------------------------------------------------------------------- |
| 1.1. Deploy at least three oracles on testnet nodes.                | ✅                |                                                                         |
| 1.2. Check the positive scenario                                    |                   |                                                                         |
| - The system successfully generates keys for ≥ 3 days.              | ✅                |                                                                         |
| - Signs Peg-out for ≥ 3 days                                        | ✅                |                                                                         |
| 1.3. Check that:                                                    |                   |                                                                         |
| - Keys are stored correctly in each oracle.                         | ✅                |                                                                         |
| - Outdated keys are pruned.                                         |                   |                                                                         |
| - Nonces and commitments are deleted after signing peg-out          | ⚠️                | Artifacts are deleted only before the start of the next peg-out signing |
| 1.4. Ensure that:                                                   |                   |                                                                         |
| - The oracle recovers upon reboot in the middle of an epoch.        |                   |                                                                         |
| - The oracle updates correctly.                                     | ✅                |                                                                         |

2. Fault tolerance and compromise resistance testing

| Description                                                                                         | Status            | Comment |
| :-------------------------------------------------------------------------------------------------- | :---------------- | :------ |
| 2.1. Turn off one or more oracles after committing to DKG:                                          | :white_check_mark: |         |
| - Ensure the system completes DKG without them                                                      |                   |         |
| 2.2. Modify one of the oracles to send incorrect packets in different rounds:                       |                   |         |
| - Garbage in round 1 packet                                                                         |                   |         |
| - Garbage in round 2 packet                                                                         |                   |         |
| - Garbage in round 3 packet                                                                         |                   |         |
| Expected result: oracles eject the bad oracle and the DKG process restarts without it               |                   |         |
| 2.3. Repeat steps 2.1 and 2.2 for the signing process.                                              |                   |         |
| 2.4. Check recovery from MTC backup.                                                                |                   |         |

3. Configurator (voting)

| Description                                                                    | Status            | Comment |
| :----------------------------------------------------------------------------- | :---------------- | :------ |
| 3.1. Add voting for:                                                           |                   |         |
| - launch                                                                       |                   |         |
| - code change (one or two contracts at once)                                   |                   |         |
| - system state change                                                          |                   |         |
| Note: this is implemented through a single voting mechanism.                   |                   |         |
| 3.2. Check the system stop/start scenario.                                     |                   |         |

4. Scalability testing

| Description                                                                    | Status            | Comment |
| :----------------------------------------------------------------------------- | :---------------- | :------ |
| 4.1. Launch the system with the maximum number of oracles.                     |                   |         |
| 4.2. Repeat all the above tests.                                               |                   |         |
| 4.3. Move oracles to public sandbox.                                           |                   |         |
| 4.4. Enable the governance contract.                                           |                   |         |

5. Slashing
  •  Requires further detailing.