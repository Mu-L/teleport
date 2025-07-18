/**
 * Teleport
 * Copyright (C) 2024 Gravitational, Inc.
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

import React, { ReactNode, useState } from 'react';
import styled from 'styled-components';

import {
  ButtonPrimary,
  ButtonSecondary,
  Flex,
  Menu,
  MenuItem,
  Text,
} from 'design';
import { CheckboxInput } from 'design/Checkbox';
import { ChevronDown } from 'design/Icon';
import { HoverTooltip } from 'design/Tooltip';

type Option<T> = {
  value: T;
  label: string | ReactNode;
  disabled?: boolean;
  disabledTooltip?: string;
};

type MultiselectMenuProps<T extends readonly Option<any>[]> = {
  options: T;
  selected: Extract<T[number], Option<any>>['value'][];
  onChange: (selected: Extract<T[number], Option<any>>['value'][]) => void;
  label: string | ReactNode;
  tooltip: string;
  /**
   * If true, renders inner control buttons (eg: apply, cancel),
   * and changes made to dropdown don't take affect until user
   * explicitly clicks on these inner control buttons.
   *
   * If false, no inner control buttons are rendered and changes
   * take affect immediately.
   */
  buffered?: boolean;
  showIndicator?: boolean;
  showSelectControls?: boolean;
  /**
   * If true, disables the button that
   * opens the dropdown menu.
   */
  disabled?: boolean;
};

export const MultiselectMenu = <T extends readonly Option<any>[]>({
  onChange,
  options,
  selected,
  label,
  tooltip,
  buffered = false,
  showIndicator = true,
  showSelectControls = true,
  disabled = false,
}: MultiselectMenuProps<T>) => {
  type Value = Extract<T[number], Option<any>>['value'];

  // we have a separate state in the filter so we can select a few different things and then click "apply"
  const [intSelected, setIntSelected] = useState<Value[]>([]);
  const [anchorEl, setAnchorEl] = useState<HTMLElement>(null);
  const handleOpen = (
    event: React.MouseEvent<HTMLButtonElement, MouseEvent>
  ) => {
    setIntSelected(selected || []);
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  // if we cancel, we reset the options to what is already selected in the params
  const cancelUpdate = () => {
    setIntSelected(selected || []);
    handleClose();
  };

  const handleSelect = (value: Value) => {
    let newSelected = (buffered ? intSelected : selected).slice();

    if (newSelected.includes(value)) {
      newSelected = newSelected.filter(v => v !== value);
    } else {
      newSelected.push(value);
    }

    (buffered ? setIntSelected : onChange)(newSelected);
  };

  const handleSelectAll = () => {
    (buffered ? setIntSelected : onChange)(
      options.filter(o => !o.disabled).map(o => o.value)
    );
  };

  const handleClearAll = () => {
    (buffered ? setIntSelected : onChange)([]);
  };

  const applyFilters = () => {
    onChange(intSelected);
    handleClose();
  };

  return (
    <Flex textAlign="center" alignItems="center">
      <HoverTooltip tipContent={tooltip}>
        <ButtonSecondary
          size="small"
          onClick={handleOpen}
          aria-haspopup="true"
          aria-expanded={!!anchorEl}
          disabled={disabled}
        >
          {label} {selected?.length > 0 ? `(${selected?.length})` : ''}
          <ChevronDown
            ml={2}
            size="small"
            color={disabled ? 'text.disabled' : 'text.slightlyMuted'}
          />
          {selected?.length > 0 && showIndicator && <FiltersExistIndicator />}
        </ButtonSecondary>
      </HoverTooltip>
      <Menu
        popoverCss={() => `margin-top: 36px;`}
        menuListCss={() => `overflow-y: auto;`}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'left',
        }}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={cancelUpdate}
      >
        {showSelectControls && (
          <MultiselectMenuOptionsContainer gap={2} p={2} position="top">
            <ButtonSecondary
              size="small"
              onClick={handleSelectAll}
              textTransform="none"
              css={`
                background-color: transparent;
              `}
              px={2}
            >
              Select All
            </ButtonSecondary>
            <ButtonSecondary
              size="small"
              onClick={handleClearAll}
              textTransform="none"
              css={`
                background-color: transparent;
              `}
              px={2}
            >
              Clear All
            </ButtonSecondary>
          </MultiselectMenuOptionsContainer>
        )}
        {options.map(opt => {
          const $checkbox = (
            <>
              <CheckboxInput
                type="checkbox"
                name={opt.value}
                disabled={opt.disabled}
                onChange={() => {
                  handleSelect(opt.value);
                }}
                id={opt.value}
                checked={(buffered ? intSelected : selected)?.includes(
                  opt.value
                )}
              />
              <Text ml={2} fontWeight={300} fontSize={2}>
                {opt.label}
              </Text>
            </>
          );
          return (
            <HoverTooltip
              key={opt.value}
              tipContent={(opt.disabled && opt.disabledTooltip) || undefined}
            >
              <MenuItem
                disabled={opt.disabled}
                px={2}
                onClick={() => (!opt.disabled ? handleSelect(opt.value) : null)}
              >
                {$checkbox}
              </MenuItem>
            </HoverTooltip>
          );
        })}
        {buffered && (
          <MultiselectMenuOptionsContainer
            justifyContent="space-between"
            p={2}
            gap={2}
            position="bottom"
          >
            <ButtonPrimary size="small" onClick={applyFilters}>
              Apply Filters
            </ButtonPrimary>
            <ButtonSecondary
              size="small"
              css={`
                background-color: transparent;
              `}
              onClick={cancelUpdate}
            >
              Cancel
            </ButtonSecondary>
          </MultiselectMenuOptionsContainer>
        )}
      </Menu>
    </Flex>
  );
};

const MultiselectMenuOptionsContainer = styled(Flex)<{
  position: 'top' | 'bottom';
}>`
  position: sticky;
  ${p => (p.position === 'top' ? 'top: 0;' : 'bottom: 0;')}
  background-color: ${p => p.theme.colors.levels.elevated};
  z-index: 1;
`;

const FiltersExistIndicator = styled.div`
  position: absolute;
  top: -4px;
  right: -4px;
  height: 12px;
  width: 12px;
  background-color: ${p => p.theme.colors.brand};
  border-radius: 50%;
  display: inline-block;
`;
