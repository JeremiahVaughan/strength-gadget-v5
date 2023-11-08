import { render } from '@testing-library/react';

import IconCogWheel from './icon-cog-wheel';

describe('IconCogWheel', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<IconCogWheel cogwheelStyle={'top'} />);
    expect(baseElement).toBeTruthy();
  });
});
