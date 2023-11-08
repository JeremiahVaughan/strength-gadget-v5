import { render } from '@testing-library/react';

import CoolButton from './cool-button';

describe('CoolButton', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<CoolButton />);
    expect(baseElement).toBeTruthy();
  });
});
