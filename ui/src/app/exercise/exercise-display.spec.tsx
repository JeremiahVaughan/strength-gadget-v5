import { render } from '@testing-library/react';

import Exercise from './exercise';

describe('Exercise', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<Exercise />);
    expect(baseElement).toBeTruthy();
  });
});
