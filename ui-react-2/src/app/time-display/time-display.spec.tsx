import { render } from '@testing-library/react';

import TimeDisplay from './time-display';

describe('TimeDisplay', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<TimeDisplay />);
    expect(baseElement).toBeTruthy();
  });
});
