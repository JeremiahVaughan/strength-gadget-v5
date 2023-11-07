import { render } from '@testing-library/react';

import MeasurementPounds from './measurement-pounds';

describe('MeasurementPounds', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<MeasurementPounds />);
    expect(baseElement).toBeTruthy();
  });
});
