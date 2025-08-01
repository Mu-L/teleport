/**
 * Teleport
 * Copyright (C) 2025  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import { Link as RouterLink } from 'react-router-dom';

import { Alert } from 'design/Alert';
import Box from 'design/Box';
import { ButtonPrimary, ButtonSecondary } from 'design/Button';
import Flex from 'design/Flex';
import { ArrowBack } from 'design/Icon';
import { Indicator } from 'design/Indicator';
import { H1 } from 'design/Text';
import { InfoGuideButton } from 'shared/components/SlidingSidePanel/InfoGuide';
import TextEditor from 'shared/components/TextEditor';
import { Attempt } from 'shared/hooks/useAsync';

import { FeatureBox, FeatureHeaderTitle } from 'teleport/components/Layout';

import { InfoGuide } from '../AuthConnectors';

/**
 * AuthConnectorEditorContent is a the content of an Auth Connector editor page.
 */
export function AuthConnectorEditorContent({
  title,
  content,
  backButtonRoute,
  isSaveDisabled,
  saveAttempt,
  fetchAttempt,
  onSave,
  onCancel,
  setContent,
  isGithub,
}: Props) {
  return (
    <FeatureBox>
      <FeatureHeaderTitle py={3} mb={2}>
        <Flex alignItems="center" justifyContent="space-between">
          <Flex alignItems="center">
            <ArrowBack
              as={RouterLink}
              mr={2}
              size="large"
              color="text.main"
              to={backButtonRoute}
            />
            <Box mr={4}>
              <H1>{title}</H1>
            </Box>
          </Flex>
          <InfoGuideButton
            config={{ guide: <InfoGuide isGitHub={isGithub} /> }}
          />
        </Flex>
      </FeatureHeaderTitle>
      {fetchAttempt.status === 'error' && (
        <Alert>{fetchAttempt.statusText}</Alert>
      )}
      {fetchAttempt.status === 'processing' && (
        <Flex alignItems="center" justifyContent="center">
          <Indicator />
        </Flex>
      )}
      {fetchAttempt.status === 'success' && (
        <Flex width="100%" height="100%">
          <Flex
            alignItems="start"
            flexDirection={'column'}
            height="100%"
            flex={4}
          >
            {saveAttempt.status === 'error' && (
              <Alert width="100%">{saveAttempt.statusText}</Alert>
            )}
            <Flex height="100%" width="100%">
              <TextEditor
                bg="levels.deep"
                readOnly={false}
                data={[{ content, type: 'yaml' }]}
                onChange={setContent}
              />
            </Flex>
            <Box mt={3}>
              <ButtonPrimary disabled={isSaveDisabled} onClick={onSave} mr="3">
                Save Changes
              </ButtonPrimary>
              <ButtonSecondary
                disabled={saveAttempt.status === 'processing'}
                onClick={onCancel}
              >
                Cancel
              </ButtonSecondary>
            </Box>
          </Flex>
        </Flex>
      )}
    </FeatureBox>
  );
}

type Props = {
  title: string;
  content: string;
  backButtonRoute: string;
  isSaveDisabled: boolean;
  saveAttempt: Attempt<void>;
  fetchAttempt: Attempt<void>;
  onSave: () => void;
  onCancel: () => void;
  setContent: (content: string) => void;
  isGithub?: boolean;
};
