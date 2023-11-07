import { render } from '@testing-library/react';

import HotButton from './hot-button';

describe('HotButton', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<HotButton />);
    expect(baseElement).toBeTruthy();
  });
});
