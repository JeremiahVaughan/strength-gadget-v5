import { render } from '@testing-library/react';

import TextBox from './text-box';

describe('TextBox', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<TextBox />);
    expect(baseElement).toBeTruthy();
  });
});
