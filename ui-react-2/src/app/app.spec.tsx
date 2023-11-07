import {render} from '@testing-library/react';

import App from './app';

describe('App', () => {

  function setup() {
    const {baseElement} = render(
        <App />
    );
    return baseElement;
  }

  it('should render successfully', () => {
    const baseElement = setup();
    expect(baseElement).toBeTruthy();
  });

});
