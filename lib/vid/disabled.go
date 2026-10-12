/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package vid

// SetDisabledIntent records that the deployment with the given virtual id should stay disabled; repeating it is a no-op.
func (this *Vid) SetDisabledIntent(vid string) error {
	_, err := this.db.Exec("INSERT INTO DisabledDeployment (VirtualId) VALUES ($1) ON CONFLICT DO NOTHING;", vid)
	return err
}

func (this *Vid) RemoveDisabledIntent(vid string) error {
	_, err := this.db.Exec("DELETE FROM DisabledDeployment WHERE VirtualId = $1;", vid)
	return err
}

func (this *Vid) HasDisabledIntent(vid string) (disabled bool, err error) {
	err = this.db.QueryRow("SELECT EXISTS (SELECT 1 FROM DisabledDeployment WHERE VirtualId = $1);", vid).Scan(&disabled)
	return disabled, err
}
