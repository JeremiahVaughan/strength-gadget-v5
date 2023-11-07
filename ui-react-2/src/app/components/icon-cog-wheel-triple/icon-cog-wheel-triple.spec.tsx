import { render } from '@testing-library/react';

import IconCogWheelTriple from './icon-cog-wheel-triple';

describe('IconCogWheelTriple', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<IconCogWheelTriple />);
    expect(baseElement).toBeTruthy();
  });
});
