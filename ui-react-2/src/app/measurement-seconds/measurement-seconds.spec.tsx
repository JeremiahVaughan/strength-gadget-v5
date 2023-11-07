import { render } from '@testing-library/react';

import MeasurementSeconds from './measurement-seconds';

describe('MeasurementSeconds', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<MeasurementSeconds />);
    expect(baseElement).toBeTruthy();
  });
});
